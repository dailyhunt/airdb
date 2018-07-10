package raft

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/etcdserver/stats"
	"github.com/coreos/etcd/pkg/types"
	"github.com/coreos/etcd/raft"
	"github.com/coreos/etcd/raft/raftpb"
	"github.com/coreos/etcd/rafthttp"
	"github.com/coreos/etcd/wal"
	"github.com/coreos/etcd/wal/walpb"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

type RNodeOptions struct {
	CurrentNodeId  uint64
	RaftPort       int
	Peers          []string
	Join           bool
	WalDir         string
	TickerInMillis int64
}

type RNode struct {
	opts             RNodeOptions
	mutationStream   <-chan []byte
	commitStream     chan []byte
	confChangeStream <-chan raftpb.ConfChange
	errorStream      chan error

	wal                 *wal.WAL
	transport           *rafthttp.Transport
	raftStorage         *raft.MemoryStorage
	stopMutationStream  chan struct{} // signals proposal channel closed
	raftHttpErrorStream chan struct{} // signals http server to shutdown
	raftHttpDoneStream  chan struct{} // signals http server shutdown complete

	logStartIndex uint64
	raft          raft.Node
	confState     raftpb.ConfState

	// index of log at start
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

func (rn *RNode) GetCommitStream() <-chan []byte {
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
	log.Info("Starting raft ........")

	oldWalExists := wal.Exist(rn.opts.WalDir)
	rn.wal = rn.replayWAL()
	log.Info("replayed WAL ........")

	peerIds := rn.preparePeerIds()
	raftCfg := rn.createRaftConfig()

	if oldWalExists {
		rn.raft = raft.RestartNode(raftCfg)
		log.Info("restarted raft node from old wal... ")
	} else {
		// Start Raft with predefined peers
		peersToStartWith := peerIds
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

	//Todo : Load Snapshot if any
	// load snapshot

	w := rn.openWal(nil)
	_, hardState, raftEntries, err := w.ReadAll()
	if err != nil {
		log.Fatalf("failed to read WAL (%v)", err)
	}

	rn.raftStorage = raft.NewMemoryStorage()
	// Todo: Load snapshot
	/*if snapshot != nil {
		rn.raftStorage.ApplySnapshot(*snapshot)
	}*/

	rn.raftStorage.SetHardState(hardState)
	// append to storage so raft starts at the right place in log
	rn.raftStorage.Append(raftEntries)

	// send nil once lastIndex is published so client knows commit channel is current
	if len(raftEntries) > 0 {
		rn.logStartIndex = raftEntries[len(raftEntries)-1].Index
	} else {
		rn.commitStream <- nil
	}

	return w

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

// Todo : Needed more customization
func (rn *RNode) preparePeerIds() []raft.Peer {
	peers := make([]raft.Peer, len(rn.opts.Peers))
	for i := range peers {
		id := uint64(i + 1)
		peers[i] = raft.Peer{ID: id}
	}
	return peers
}

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
	peers := rn.opts.Peers
	for i := range peers {
		peerId := types.ID(i + 1) // Adding id to peers based on index basis on comma separated string
		us := []string{peers[i]}
		rn.transport.AddPeer(peerId, us)

	}
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

			/*if !raft.IsEmptySnap(rd.Snapshot) {
				rn.saveSnap(rd.Snapshot)
				rn.raftStorage.ApplySnapshot(rd.Snapshot)
				rn.publishSnapshot(rd.Snapshot)
			}*/
			rn.raftStorage.Append(rd.Entries)
			rn.transport.Send(rd.Messages)
			entriesToApply := rn.entriesToApply(rd.CommittedEntries)
			if ok := rn.publishEntries(entriesToApply); !ok {
				rn.stop()
				return
			}
			/*rn.maybeTriggerSnapshot()*/ // TODO
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
func (rn *RNode) initializeSnapshotMeta() {
	/*snap, err := rc.raftStorage.Snapshot()
	if err != nil {
		panic(err)
	}
	rc.confState = snap.Metadata.ConfState
	rc.snapshotIndex = snap.Metadata.Index
	rc.appliedIndex = snap.Metadata.Index*/
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
	// Todo ::
	/*if len(entries) == 0 {
		return
	}
	firstIdx := entries[0].Index
	if firstIdx > rn.appliedIndex+1 {
		log.Fatalf("first index of committed entry[%d] should <= progress.appliedIndex[%d]+1", firstIdx, rc.appliedIndex)
	}
	if rn.appliedIndex-firstIdx+1 < uint64(len(entries)) {
		toApply = entries[rn.appliedIndex-firstIdx+1:]
	}
	return toApply*/
	return entries
}
func (rn *RNode) publishEntries(entries []raftpb.Entry) bool {
	for i := range entries {

		entry := entries[i]
		switch entry.Type {

		case raftpb.EntryNormal:
			if len(entry.Data) == 0 {
				// ignore empty messages
				break
			}
			s := string(entry.Data)
			fmt.Println("Normal entry ", s)
			select {
			case rn.commitStream <- entry.Data:
			case <-rn.mutationStream:
				return false
			}

		case raftpb.EntryConfChange:
			var cc raftpb.ConfChange
			cc.Unmarshal(entry.Data)
			rn.confState = *rn.raft.ApplyConfChange(cc)
			switch cc.Type {
			// Todo : Need to doc and modify
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

		// after commit, update appliedIndex
		// Todo ::
		/*rn.appliedIndex = entry.Index*/

		// special nil commit to signal replay has finished
		// Todo ::
		/*if entry.Index == rn.lastIndex {
			select {
			case rn.commitStream <- nil:
			case <-rn.mutationStream:
				return false
			}
		}*/
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
		commitStream:        make(chan []byte),
		confChangeStream:    confChangeStream,
		stopMutationStream:  make(chan struct{}),
		raftHttpErrorStream: make(chan struct{}),
		raftHttpDoneStream:  make(chan struct{}),
	}

	go node.startRaft()
	return node
}
