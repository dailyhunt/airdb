package raft

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/etcdserver/stats"
	"github.com/coreos/etcd/pkg/fileutil"
	"github.com/coreos/etcd/pkg/types"
	"github.com/coreos/etcd/raft"
	"github.com/coreos/etcd/raft/raftpb"
	"github.com/coreos/etcd/rafthttp"
	"github.com/coreos/etcd/snap"
	"github.com/coreos/etcd/wal"
	"github.com/coreos/etcd/wal/walpb"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

var defaultSnapshotCount uint64 = 10000

type RNodeOptions struct {
	CurrentNodeId  uint64
	RaftPort       int
	Join           bool
	WalDir         string
	SnapshotDir    string
	TickerInMillis int64
}

type RNode struct {
	opts             RNodeOptions
	mutationStream   <-chan []byte
	commitStream     chan raftpb.Entry
	confChangeStream <-chan raftpb.ConfChange
	errorStream      chan error

	wal                     *wal.WAL
	transport               *rafthttp.Transport
	raftStorage             *raft.MemoryStorage
	raft                    raft.Node
	raftConfigState         raftpb.ConfState
	lastLogIndex            uint64
	lastCommittedEntryIndex uint64

	stopMutationStream  chan struct{} // signals proposal channel closed
	raftHttpErrorStream chan struct{} // signals http server to shutdown
	raftHttpDoneStream  chan struct{}
	snapshotter         *snap.Snapshotter
	snapshotterReady    chan<- *snap.Snapshotter
	snapshotIndex       uint64
	snapshotCount       uint64

	// signals http server shutdown complete
}

func (rn *RNode) Process(ctx context.Context, m raftpb.Message) error {
	return rn.raft.Step(ctx, m)
}

func (rn *RNode) IsIDRemoved(id uint64) bool {
	return false
}

func (rn *RNode) ReportUnreachable(id uint64) {

}

func (rn *RNode) ReportSnapshot(id uint64, status raft.SnapshotStatus) {

}

func (rn *RNode) GetCommitStream() <-chan raftpb.Entry {
	return rn.commitStream
}

func (rn *RNode) GetErrorStream() <-chan error {
	return rn.errorStream
}

func (rn *RNode) startRaft() {
	rn.startRaftWithWal()
	rn.initializeRaftTransport()
	go rn.serveRaft()
	go rn.serveChannels()
}

func (rn *RNode) startRaftWithWal() {
	log.Info("starting raft ........")

	// Create snapshotter meta/files
	snapDir := rn.opts.SnapshotDir
	if !fileutil.Exist(snapDir) {
		if err := os.Mkdir(snapDir, 0750); err != nil {
			log.Fatalf("cannot create dir for snapshot (%v)", err)
		}
	}
	rn.snapshotter = snap.New(snapDir)
	rn.snapshotterReady <- rn.snapshotter

	oldWalExists := wal.Exist(rn.opts.WalDir)
	rn.wal = rn.replayWAL()
	log.Info("replayed WAL ........")

	//peerIds := rn.preparePeerIds()
	raftCfg := rn.createRaftConfig()

	if oldWalExists {
		rn.raft = raft.RestartNode(raftCfg)
		log.Info("restarted raft node from old wal... ")
	} else {
		//Not Supported - Start Raft with predefined peers
		// Will always single node and add to cluster
		peersToStartWith := []raft.Peer{{ID: rn.opts.CurrentNodeId}}
		// Otherwise add node to peers first and start with join flag
		if rn.opts.Join {
			peersToStartWith = nil
		}
		rn.raft = raft.StartNode(raftCfg, peersToStartWith)
		log.Info("started raft node with new wal")
	}
}

func (rn *RNode) replayWAL() *wal.WAL {
	log.Infof("replaying WAL of member %d at dir %s", rn.opts.CurrentNodeId, rn.opts.WalDir)

	snapshot := rn.loadSnapshot()
	w := rn.openWal(snapshot)
	_, hardState, raftEntries, err := w.ReadAll()

	// TODO: TO Remove
	if len(raftEntries) > 0 {
		log.Debugf("last entry index %d first entry index %d ", raftEntries[len(raftEntries)-1].Index, raftEntries[0].Index)
	}

	if err != nil {
		log.Fatalf("failed to read WAL (%v)", err)
	}

	rn.raftStorage = raft.NewMemoryStorage()
	if snapshot != nil {
		rn.raftStorage.ApplySnapshot(*snapshot)
	}

	rn.raftStorage.SetHardState(hardState)
	// append to storage so raft starts at the right place in log
	rn.raftStorage.Append(raftEntries)
	// send nil once lastIndex is published so client knows commit channel is current
	if len(raftEntries) > 0 {
		// this is needed to mark if replayed entries are published to raft in publishEntries
		rn.lastLogIndex = raftEntries[len(raftEntries)-1].Index
	} else {
		rn.commitStream <- raftpb.Entry{}
	}

	return w
}

func (rn *RNode) loadSnapshot() *raftpb.Snapshot {
	snapshot, err := rn.snapshotter.Load()
	if err != nil && err != snap.ErrNoSnapshot {
		log.Fatalf("error loading snapshot (%v)", err)
	}
	return snapshot
}

func (rn *RNode) openWal(lastSnapshot *raftpb.Snapshot) *wal.WAL {
	walDir := rn.opts.WalDir

	// Create wal if not exist
	if !wal.Exist(walDir) {
		if err := os.Mkdir(walDir, 0750); err != nil {
			log.Fatalf("cannot create wal dir at %s for node %d due to (%v)", walDir, rn.opts.CurrentNodeId, err)
		}

		w, err := wal.Create(walDir, nil)
		if err != nil {
			log.Fatalf("cannot create wal file at %s for node %d due to (%v)", walDir, rn.opts.CurrentNodeId, err)
		}
		w.Close()
	}

	snapshotToStart := walpb.Snapshot{}

	// update last saved snapshot
	if lastSnapshot != nil {
		snapshotToStart.Index = lastSnapshot.Metadata.Index
		snapshotToStart.Term = lastSnapshot.Metadata.Term
	}
	log.Infof("loading WAL at term %d and index %d", snapshotToStart.Term, snapshotToStart.Index)

	raftWal, err := wal.Open(walDir, snapshotToStart)
	if err != nil {
		log.Fatalf("error while loading wal for node id %s due to (%v)", rn.opts.CurrentNodeId, err)
	}

	return raftWal

}

// Todo: config from File or options
func (rn *RNode) createRaftConfig() *raft.Config {
	return &raft.Config{
		ID:            rn.opts.CurrentNodeId,
		ElectionTick:  10,
		HeartbeatTick: 1,
		Storage:       rn.raftStorage,
		//Applied:                   0,
		MaxSizePerMsg:   1024 * 1024,
		MaxInflightMsgs: 256,
		CheckQuorum:     false,
		PreVote:         false,
		//ReadOnlyOption:  0,
		//Logger:          nil,
		//DisableProposalForwarding: false,
	}

}

// Todo :: Related to stats and cluster id
// Only start itself. All information related to peers will be
// added through api
func (rn *RNode) initializeRaftTransport() {
	id := rn.opts.CurrentNodeId
	rn.transport = &rafthttp.Transport{
		ID:          types.ID(id),
		ClusterID:   0x1000,
		Raft:        rn,
		ServerStats: stats.NewServerStats("", ""),
		LeaderStats: stats.NewLeaderStats(strconv.FormatUint(id, 10)),
		ErrorC:      make(chan error),
	}

	rn.transport.Start()
	/*		peers := rn.opts.Peers
			for i := range peers {
				peerId := types.ID(i + 1) // Adding id to peers based on index basis on comma separated string
				us := []string{peers[i]}
				rn.transport.AddPeer(peerId, us)

			}*/
	log.Info("started raft transport .....")

}

// Setting up raft server over http
func (rn *RNode) serveRaft() {
	raftUrlStr := fmt.Sprintf("http://localhost:%v", rn.opts.RaftPort)
	log.Infof("starting raft http server ar %s ", raftUrlStr)
	raftUrl, err := url.Parse(raftUrlStr)
	if err != nil {
		log.Fatalf("failed parsing raftUrl (%v)", err)
	}

	listener, err := newRaftListener(raftUrl.Host, rn.raftHttpErrorStream)
	if err != nil {
		log.Fatalf("failed to listen airdb raft http (%v)", err)
	}

	err = (&http.Server{Handler: rn.transport.Handler()}).Serve(listener)
	select {
	case <-rn.raftHttpErrorStream:
	default:
		log.Fatalf("failed to serve raft http (%v)", err)
	}

	close(rn.raftHttpDoneStream)

}

func (rn *RNode) serveChannels() {
	rn.initializeSnapshotMeta()
	go rn.startProposingMutation()
	defer rn.wal.Close()
	ticker := time.NewTicker(time.Duration(rn.opts.TickerInMillis) * time.Millisecond)

	// event loop on raft state machine updates
	for {
		select {
		case <-ticker.C:
			rn.raft.Tick()
		case rd := <-rn.raft.Ready():
			rn.wal.Save(rd.HardState, rd.Entries)

			if len(rd.Entries) > 10 && log.GetLevel() == log.DebugLevel {
				log.Debugf("Total entries %d :: Hard State %s ", len(rd.Entries), rd.HardState.String())
				log.Debugf("Entries %v ", EntriesToString(rd.Entries))

			}

			// TODO :: Need to test
			if !raft.IsEmptySnap(rd.Snapshot) {
				rn.saveSnap(rd.Snapshot)
				rn.raftStorage.ApplySnapshot(rd.Snapshot)
				rn.publishSnapshot(rd.Snapshot)
			}
			rn.raftStorage.Append(rd.Entries)
			rn.transport.Send(rd.Messages)
			entriesToApply := rn.entriesToApply(rd.CommittedEntries)
			if ok := rn.publishEntries(entriesToApply); !ok {
				rn.stop()
				return
			}

			rn.raft.Advance()

		case err := <-rn.transport.ErrorC:
			rn.writeError(err)
			return
		case <-rn.stopMutationStream:
			rn.stop()
			return
		}
	}

}

// Todo :: Need to verify yet
func (rn *RNode) publishSnapshot(snapshotToSave raftpb.Snapshot) {
	if raft.IsEmptySnap(snapshotToSave) {
		return
	}

	log.Infof("publishing snapshot at index %d", rn.snapshotIndex)
	defer log.Infof("finished publishing snapshot at index %d", rn.snapshotIndex)

	if snapshotToSave.Metadata.Index <= rn.lastCommittedEntryIndex {
		log.Fatalf("snapshot index [%d] should > progress.appliedIndex [%d] + 1", snapshotToSave.Metadata.Index, rn.lastCommittedEntryIndex)
	}

	rn.commitStream <- raftpb.Entry{} // trigger fsm to load snapshot

	rn.raftConfigState = snapshotToSave.Metadata.ConfState
	rn.snapshotIndex = snapshotToSave.Metadata.Index
	rn.lastCommittedEntryIndex = snapshotToSave.Metadata.Index
}

func (rn *RNode) MaybeTriggerSnapshot(fromCommittedIndex uint64) {
	log.Infof("start snapshot [last committed applied index: %d | last snapshot index: %d]", fromCommittedIndex, rn.snapshotIndex)

	// following code is not needed as we are triggering snapshot whenever
	// memtable is flushed .
	/*if fromCommittedIndex-rn.snapshotIndex <= rn.snapshotCount {
		return
	}*/

	snapshot, err := rn.raftStorage.CreateSnapshot(fromCommittedIndex, &rn.raftConfigState, nil)
	if err != nil {
		panic(err)
	}
	if err := rn.saveSnap(snapshot); err != nil {
		panic(err)
	}

	if err := rn.raftStorage.Compact(fromCommittedIndex); err != nil {
		panic(err)
	}

	log.Printf("compacted log at index %d", fromCommittedIndex)
	rn.snapshotIndex = fromCommittedIndex

}

func (rn *RNode) saveSnap(snap raftpb.Snapshot) error {
	// must save the snapshot index to the WAL before saving the
	// snapshot to maintain the invariant that we only Open the
	// wal at previously-saved snapshot indexes.
	walSnap := walpb.Snapshot{
		Index: snap.Metadata.Index,
		Term:  snap.Metadata.Term,
	}
	if err := rn.wal.SaveSnapshot(walSnap); err != nil {
		return err
	}
	if err := rn.snapshotter.SaveSnap(snap); err != nil {
		return err
	}
	return rn.wal.ReleaseLockTo(snap.Metadata.Index)
}

// Initialize snapshot meta including lastCommittedEntryIndex
func (rn *RNode) initializeSnapshotMeta() {
	snap, err := rn.raftStorage.Snapshot()
	if err != nil {
		panic(err)
	}
	rn.raftConfigState = snap.Metadata.ConfState
	rn.snapshotIndex = snap.Metadata.Index
	rn.lastCommittedEntryIndex = snap.Metadata.Index
}
func (rn *RNode) startProposingMutation() {
	var confChangeCount uint64 = 0
	for rn.mutationStream != nil && rn.confChangeStream != nil {
		select {
		case mutation, ok := <-rn.mutationStream:
			if !ok {
				rn.mutationStream = nil
			} else {
				rn.raft.Propose(context.TODO(), mutation)
			}
		case confChange, ok := <-rn.confChangeStream:
			if !ok {
				rn.confChangeStream = nil
			} else {
				confChangeCount += 1
				confChange.ID = confChangeCount
				rn.raft.ProposeConfChange(context.TODO(), confChange)
			}
		}
	}

	close(rn.stopMutationStream)

}

func (rn *RNode) entriesToApply(entries []raftpb.Entry) (toApply []raftpb.Entry) {
	if len(entries) == 0 {
		return
	}
	firstIdx := entries[0].Index

	if firstIdx > rn.lastCommittedEntryIndex+1 {
		log.Fatalf("first index of committed entry[%d] should <= progress.appliedIndex[%d]+1", firstIdx, rn.lastCommittedEntryIndex)
	}

	indexDiff := rn.lastCommittedEntryIndex - firstIdx
	startIndex := indexDiff + 1

	if startIndex < uint64(len(entries)) {
		toApply = entries[startIndex:]
	}
	return toApply
}

// publishEntries writes committed log entries to commit channel and returns
// whether all entries could be published.
func (rn *RNode) publishEntries(entries []raftpb.Entry) bool {

	for i := range entries {
		entry := entries[i]

		switch entry.Type {

		case raftpb.EntryNormal:
			if len(entry.Data) == 0 {
				break // ignore empty messages
			}
			select {
			case rn.commitStream <- entry:
			case <-rn.stopMutationStream:
				return false
			}

		case raftpb.EntryConfChange:
			var cc raftpb.ConfChange
			cc.Unmarshal(entry.Data)
			rn.raftConfigState = *rn.raft.ApplyConfChange(cc)
			switch cc.Type {
			// Todo : Need to add document and modify
			case raftpb.ConfChangeAddNode:
				if len(cc.Context) > 0 {
					rn.transport.AddPeer(types.ID(cc.NodeID), []string{string(cc.Context)})
				}
			case raftpb.ConfChangeRemoveNode:
				if cc.NodeID == uint64(rn.opts.CurrentNodeId) {
					log.Println("I've been removed from the cluster! Shutting down.")
					return false
				}
				rn.transport.RemovePeer(types.ID(cc.NodeID))
			}
		}

		// after commit, update last appliedIndex.
		// this wll be need to ignore old entries or entries which are already applied
		rn.lastCommittedEntryIndex = entry.Index

		// special nil commit to signal replay has finished
		if entry.Index == rn.lastLogIndex {
			select {
			case rn.commitStream <- raftpb.Entry{}:
			case <-rn.stopMutationStream:
				return false
			}
		}
	}
	return true
}

func (rn *RNode) writeError(err error) {
	rn.stopHTTP()
	close(rn.commitStream)
	rn.errorStream <- err
	close(rn.errorStream)
	rn.raft.Stop()
}

func (rn *RNode) stop() {
	rn.stopHTTP()
	close(rn.commitStream)
	close(rn.errorStream)
	rn.raft.Stop()
}

func (rn *RNode) stopHTTP() {
	rn.transport.Stop()
	close(rn.raftHttpErrorStream)
	<-rn.raftHttpDoneStream
}

func NewRaftNode(opts RNodeOptions, mutationStream <-chan []byte, confChangeStream <-chan raftpb.ConfChange) *RNode {

	node := &RNode{
		opts:                opts,
		mutationStream:      mutationStream,
		commitStream:        make(chan raftpb.Entry),
		confChangeStream:    confChangeStream,
		stopMutationStream:  make(chan struct{}),
		raftHttpErrorStream: make(chan struct{}),
		raftHttpDoneStream:  make(chan struct{}),
		snapshotterReady:    make(chan *snap.Snapshotter, 1),
		snapshotCount:       defaultSnapshotCount,
	}

	go node.startRaft()
	return node
}
