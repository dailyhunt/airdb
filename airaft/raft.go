package airaft

import (
	"github.com/coreos/etcd/etcdserver/api/rafthttp"
	"github.com/coreos/etcd/raft"
	"github.com/coreos/etcd/raft/raftpb"
	"github.com/coreos/etcd/wal"

	"context"
	"fmt"
	"github.com/coreos/etcd/etcdserver/api/snap"
	stats "github.com/coreos/etcd/etcdserver/api/v2stats"
	"github.com/coreos/etcd/pkg/fileutil"
	"github.com/coreos/etcd/pkg/types"
	"github.com/coreos/etcd/wal/walpb"
	"github.com/dailyhunt/airdb/utils"
	logger "github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

var defaultSnapshotCount uint64 = 10000

type raftNode struct {

	// TODO: (shailendra)channel should be of other type instead of string
	// TODO: Type still needed to decide
	proposeChan      <-chan string            // operation.Op        // proposed key value (Todo: Check redirection)
	raftConfChangeCh <-chan raftpb.ConfChange // proposed raft/region config change(Todo: Check redirection)
	commitLogChan    chan<- *string           // entries committed to log/wal for (key,value) (Todo: Check redirection)
	errorCh          chan<- error             // errors from raft session (Todo: Check redirection)

	// Basic Raft info
	id    int      // id for raft session
	join  bool     // flag , if node is joining existing cluster
	peers []string // url of raft peers

	// wal
	walDir string
	wal    *wal.WAL

	// snapshot (Todo)
	snapDir          string
	getSnapshot      func() ([]byte, error)
	snapshotIndex    uint64
	snapshotCount    uint64
	snapshotter      *snap.Snapshotter
	snapshotterReady chan *snap.Snapshotter // signals when snapshotter is ready

	// raft
	raftState     raftpb.ConfState
	raftNode      raft.Node
	raftStorage   *raft.MemoryStorage
	raftTransport *rafthttp.Transport

	lastAppliedIndex uint64 // Todo (doc)
	lastIndex        uint64 // index of log at start

	// Explicit cancellation of channels
	stopCh     chan struct{}
	httpStopCh chan struct{}
	httpDoneCh chan struct{}
}

// newRaftNode initiates a raft instance and returns a committed log entry
// channel and error channel. Proposals for log updates are sent over the
// provided the proposal channel. All log entries are replayed over the
// commit channel, followed by a nil message (to indicate the channel is
// current), then new log entries. To shutdown, close proposeC and read errorC.
func NewRaftNode(id int, peers []string, join bool, getSnapshot func() ([]byte, error), proposeC <-chan string,
	confChangeC <-chan raftpb.ConfChange) (<-chan *string, <-chan error, <-chan *snap.Snapshotter) {

	commitC := make(chan *string)
	errorC := make(chan error)

	rc := &raftNode{
		proposeChan:      proposeC,
		raftConfChangeCh: confChangeC,
		commitLogChan:    commitC,
		errorCh:          errorC,
		id:               id,
		peers:            peers,
		join:             join,
		walDir:           fmt.Sprintf("raftexample-%d", id),
		snapDir:          fmt.Sprintf("raftexample-%d-snap", id),
		getSnapshot:      getSnapshot,
		snapshotCount:    defaultSnapshotCount,
		stopCh:           make(chan struct{}),
		httpStopCh:       make(chan struct{}),
		httpDoneCh:       make(chan struct{}),

		snapshotterReady: make(chan *snap.Snapshotter, 1),
		// rest of structure populated after WAL replay
	}
	go rc.startRaft()
	return commitC, errorC, rc.snapshotterReady
}

func (rc *raftNode) saveSnap(snap raftpb.Snapshot) error {
	/*// must save the snapshot index to the WAL before saving the
	// snapshot to maintain the invariant that we only Open the
	// wal at previously-saved snapshot indexes.
	walSnap := walpb.Snapshot{
		Index: snap.Metadata.Index,
		Term:  snap.Metadata.Term,
	}
	if err := rc.wal.SaveSnapshot(walSnap); err != nil {
		return err
	}
	if err := rc.snapshotter.SaveSnap(snap); err != nil {
		return err
	}
	return rc.wal.ReleaseLockTo(snap.Metadata.Index)*/
	return nil
}

// Find out entries to apply/publish
// Ignore old entries from last index
func (rc *raftNode) entriesToApply(inEntries []raftpb.Entry) (outEntries []raftpb.Entry) {
	logger.Debug("Entries to apply ", len(inEntries))
	if len(inEntries) == 0 {
		return
	}

	firstIndex := inEntries[0].Index
	if firstIndex > rc.lastAppliedIndex+1 {
		logger.Fatalf("first index of committed entry[%d] should <= progress.appliedIndex[%d]+1", firstIndex, rc.lastAppliedIndex)
	}

	if rc.lastAppliedIndex-firstIndex+1 < uint64(len(inEntries)) {
		// Slice after verification
		outEntries = inEntries[rc.lastAppliedIndex-firstIndex+1:]
	}

	return outEntries

}

// Write committed logs to commit Channel
// Return if all entries could be published
func (rc *raftNode) publishEntries(entries []raftpb.Entry) bool {
	logger.Debug("publish entries of size ", len(entries))
	for i := range entries {
		entry := entries[i]
		entryType := entry.Type

		switch entryType {
		case raftpb.EntryNormal:
			if len(entry.Data) == 0 {
				break // Ignore empty entries
			}

			// Todo: Handle different operations based on type
			// Todo: Handle Type of message as of now it is string
			/*var put operation.Put
			json.Unmarshal(entry.Data, &put)*/
			st := rc.raftNode.Status()
			logger.Info(fmt.Sprintf("[ time : %s ] = I am node [ id : %d] - [term :  %d ] - [ vote : %d ] - [ commit : %d ] - "+
				"[ state : %s ] - [ applied : %v ] - [ my-leader : %d ]",
				utils.GetCurrentTime(), st.ID, st.Term, st.Commit, st.Vote, st.RaftState.String(), st.Applied, st.Lead))

			s := string(entry.Data)
			select {
			case rc.commitLogChan <- &s: // todo (pointer allocation)
			case <-rc.stopCh:
				return false
			}

		case raftpb.EntryConfChange:
			var cc raftpb.ConfChange
			cc.Unmarshal(entry.Data)
			rc.raftState = *rc.raftNode.ApplyConfChange(cc)
			switch cc.Type {
			case raftpb.ConfChangeAddNode:
				if len(cc.Context) > 0 {
					nodeId := types.ID(cc.NodeID)
					us := []string{string(cc.Context)} // Todo : Log this
					logger.Info("Adding node to cluster ", nodeId, " us -> ", us)
					rc.raftTransport.AddPeer(nodeId, us)
				}
			case raftpb.ConfChangeRemoveNode:
				if cc.NodeID == uint64(rc.id) {
					logger.Info("I've been removed from the cluster! Shutting down.")
					return false
				}
				rc.raftTransport.RemovePeer(types.ID(cc.NodeID))
			}

		} // switch

		// after commit, update appliedIndex
		rc.lastAppliedIndex = entry.Index

		// special nil commit to signal replay has finished
		// Todo : why it is needed ??
		if entry.Index == rc.lastIndex {
			logger.Debug("entry index and last index is same")
			select {
			case rc.commitLogChan <- nil:
			case <-rc.stopCh:
				return false
			}
		}
	} // for

	return true
}

// Todo - implement
func (rc *raftNode) loadSnapshot() *raftpb.Snapshot {
	logger.Debug("Loading snapshot ")
	snapshot, err := rc.snapshotter.Load()
	if err != nil && err != snap.ErrNoSnapshot {
		logger.Fatalf("airdb: error loading snapshot (%v)", err)
	}
	return snapshot
}

// Todo (take snapshot as argument)
func (rc *raftNode) openWal(snapshot *raftpb.Snapshot) *wal.WAL {
	if !wal.Exist(rc.walDir) {
		if err := os.Mkdir(rc.walDir, 0750); err != nil {
			logger.Fatalf("airdb: cannot create dir for wal (%v)", err)
		}

		w, err := wal.Create(zap.NewExample(), rc.walDir, nil)
		if err != nil {
			logger.Fatalf("airdb : create wal error (%v)", err)
		}
		w.Close()
	}

	walSnapshot := walpb.Snapshot{}

	// set term and index form existing snapshot
	if snapshot != nil {
		walSnapshot.Index, walSnapshot.Term = snapshot.Metadata.Index, snapshot.Metadata.Term
	}

	// Todo (need more understanding)
	logger.Printf("loading WAL snapshot at term %d and index %d", walSnapshot.Term, walSnapshot.Index)
	w, err := wal.Open(zap.NewExample(), rc.walDir, walSnapshot)
	if err != nil {
		logger.Fatalf("airdb: error loading wal (%v)", err)
	}

	return w

}

func (rc *raftNode) replayWal() *wal.WAL {

	// TODO(sohan) : snapshot and replay loading
	logger.Printf("replaying WAL of member %d", rc.id)
	snapshot := rc.loadSnapshot()
	w := rc.openWal(snapshot)
	_, st, ents, err := w.ReadAll()
	logger.Debug("Size of entries ", len(ents))
	if err != nil {
		logger.Fatalf("airdb: failed to read WAL (%v)", err)
	}
	rc.raftStorage = raft.NewMemoryStorage()
	if snapshot != nil {
		rc.raftStorage.ApplySnapshot(*snapshot)
	}

	logger.WithFields(logger.Fields{
		"commit": st.Commit,
		"term":   st.Term,
		"vote":   st.Vote,
	}).Debug("hard state from replaying log")
	rc.raftStorage.SetHardState(st)

	// append to storage so raft starts at the right place in log
	rc.raftStorage.Append(ents)
	// send nil once lastIndex is published so client knows commit channel is current
	if len(ents) > 0 {
		rc.lastIndex = ents[len(ents)-1].Index
	} else {
		logger.Debug("replay wal : sending nil to commit channel")
		rc.commitLogChan <- nil
	}
	return w
}

func (rc *raftNode) writeError(err error) {
	rc.stopHTTP()
	close(rc.commitLogChan)
	rc.errorCh <- err
	close(rc.errorCh)
	rc.raftNode.Stop()
}

func (rc *raftNode) startRaft() {

	// Todo(sohan): Snapshot
	if !fileutil.Exist(rc.snapDir) {
		if err := os.Mkdir(rc.snapDir, 0750); err != nil {
			logger.Fatalf("airdb: cannot create dir for snapshot (%v)", err)
		}
	}
	rc.snapshotter = snap.New(zap.NewExample(), rc.snapDir)
	rc.snapshotterReady <- rc.snapshotter

	// Todo(sohan): WAL
	oldwal := wal.Exist(rc.walDir)
	rc.wal = rc.replayWal()

	// Todo(sohan): Raft Config for peers
	rcPeers := make([]raft.Peer, len(rc.peers))
	for i := range rc.peers {
		rcPeers[i] = raft.Peer{ID: uint64(i + 1)} // Todo - doc
		logger.Info("Adding peer to raft ", rcPeers[i].ID)
	}

	// TODO(sohan) - set read only replica using learners array
	logger.Debug("Creating Raft Config for ", rc.id)

	// Todo(Sohan) - take config from Config file
	rConfig := &raft.Config{
		ID:              uint64(rc.id),
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         rc.raftStorage,
		MaxSizePerMsg:   1024 * 1024,
		MaxInflightMsgs: 256,
	}

	// Todo - Start vs restart node
	if oldwal {
		logger.Debug("Re-Starting Raft Server")
		rc.raftNode = raft.RestartNode(rConfig)
	} else {
		logger.Debug("Starting Raft Server")
		startPeers := rcPeers
		if rc.join {
			startPeers = nil
		}
		rc.raftNode = raft.StartNode(rConfig, startPeers)
	}

	transLogger, _ := zap.NewProduction()

	// Todo(sohan): Raft Transport
	rc.raftTransport = &rafthttp.Transport{
		Logger:      transLogger,
		ID:          types.ID(rc.id),
		ClusterID:   0x1000,
		Raft:        rc,
		ServerStats: stats.NewServerStats("", ""),
		LeaderStats: stats.NewLeaderStats(strconv.Itoa(rc.id)),
		ErrorC:      make(chan error),
	}

	logger.WithFields(logger.Fields{
		"clusterId":   rc.raftTransport.ClusterID,
		"transportId": rc.raftTransport.ID,
	}).Info("Starting Raft Transport")

	rc.raftTransport.Start()

	for i := range rc.peers {
		//if index+1 is not equal to peer id (node id)
		// (we are assuming that id are starting from 0),
		// then add peer to transport
		if i+1 != rc.id {
			peerId := types.ID(i + 1)
			urlStr := []string{rc.peers[i]} // url string [<host>:<port>] as singleton array
			logger.WithFields(logger.Fields{
				"id":        rc.raftTransport.ID,
				"peerId":    peerId,
				"urlStr":    urlStr,
				"clusterId": rc.raftTransport.ClusterID,
			}).Info("Adding Peer to Transport")

			rc.raftTransport.AddPeer(peerId, urlStr)
		}
	}

	go rc.serveRaft()
	go rc.serveChannels()

	logger.Debug("started raft")

}

func (rc *raftNode) stopRaft() {
	rc.stopHTTP()
	close(rc.commitLogChan)
	close(rc.errorCh)
	rc.raftNode.Stop()
}

func (rc *raftNode) stopHTTP() {
	rc.raftTransport.Stop()
	close(rc.httpStopCh)
	<-rc.httpDoneCh
}

// Todo: Doc
func (rc *raftNode) serveRaft() {
	logger.Debug("Starting Serving Raft transport", rc.id)
	transportAddress := rc.peers[rc.id-1]
	url, err := url.Parse(transportAddress)
	if err != nil {
		logger.Fatalf("Failing parsing URL for raft transport (%v)", err)
	}
	logger.Debug("Raft Transport Address ", url)
	listener, err := newRaftListener(url.Host, rc.stopCh)
	if err != nil {
		logger.Fatalf("Failed to listen for raft http : (%v)", err)
	}

	server := &http.Server{Handler: rc.raftTransport.Handler()}
	err = server.Serve(listener)

	select {
	case <-rc.httpStopCh:
	default:
		logger.Fatalf("Failed to server raft http (%v)", err)
	}

	close(rc.httpDoneCh)

}

func (rc *raftNode) serveChannels() {
	// Todo : Wal and snapshot related

	defer rc.wal.Close()

	// Todo(sohan): Take from config
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	go rc.proposeToRaft()

	// Event loop on raft state machine updates
	for {
		select {
		case <-ticker.C:
			rc.raftNode.Tick()
			//logger.Debug("Tick")

			// store raft entries to wal, then publish over commit channel
		case rd := <-rc.raftNode.Ready():
			// Todo:
			rc.wal.Save(rd.HardState, rd.Entries)
			// Todo: Save to WAL
			// Todo: Snapshot stuff
			//state := rc.raftNode.Status().RaftState.String()
			rc.raftStorage.Append(rd.Entries)
			rc.raftTransport.Send(rd.Messages)

			outEntries := rc.entriesToApply(rd.CommittedEntries)
			if ok := rc.publishEntries(outEntries); !ok {
				rc.stopRaft()
				return
			}

			// todo : Trigger snapshot
			rc.raftNode.Advance()
		case err := <-rc.raftTransport.ErrorC:
			rc.writeError(err)
			return
		case <-rc.stopCh:
			rc.stopRaft()
			return
		}

	} //select end

	logger.Debug("end of debug loop")

}

func (rc *raftNode) proposeToRaft() {
	var configChangeCount uint64 = 0 // for config change id
	logger.Debug("propose raft ")

	for rc.proposeChan != nil && rc.raftConfChangeCh != nil {
		select {
		case proposed, ok := <-rc.proposeChan:
			if !ok {
				rc.proposeChan = nil
				logger.Debug("nil value from proposed channel")
			} else {
				// Todo(sohan)- switch based on type
				//op := proposed.Name()
				logger.Debug("value from proposed channel ", proposed)
				rc.raftNode.Propose(context.TODO(), []byte(proposed))
				/*switch op {
				case "PUT":
					logger.Debug("Processing Put in proposeRaft ", proposed.String())
					rc.raftNode.Propose(context.TODO(), []byte(proposed.String()))
				default:
					// TODO
					logger.Warn("Invalid operation")
				}*/

			}

		case configChange, ok := <-rc.raftConfChangeCh:
			logger.Debug("raft config change")
			if !ok {
				rc.raftConfChangeCh = nil
			} else {
				configChangeCount += 1
				configChange.ID = configChangeCount
				// Todo : check for why use context.TODO()
				rc.raftNode.ProposeConfChange(context.TODO(), configChange)
			}
		} // select end
	} // for end

	// client closed channel; shutdown raft if not already
	close(rc.stopCh)

}

func (rc *raftNode) Process(ctx context.Context, m raftpb.Message) error {
	return rc.raftNode.Step(ctx, m)
}

func (rc *raftNode) IsIDRemoved(id uint64) bool {
	return false
}

func (rc *raftNode) ReportUnreachable(id uint64) {
	//panic("implement me")
}

func (rc *raftNode) ReportSnapshot(id uint64, status raft.SnapshotStatus) {
	//panic("implement me")
}

// Todo
func (rc *raftNode) publishSnapshot(snapshotToSave raftpb.Snapshot) {
	/*if raft.IsEmptySnap(snapshotToSave) {
		return
	}

	log.Printf("publishing snapshot at index %d", rc.snapshotIndex)
	defer log.Printf("finished publishing snapshot at index %d", rc.snapshotIndex)

	if snapshotToSave.Metadata.Index <= rc.appliedIndex {
		log.Fatalf("snapshot index [%d] should > progress.appliedIndex [%d] + 1", snapshotToSave.Metadata.Index, rc.appliedIndex)
	}
	rc.commitC <- nil // trigger kvstore to load snapshot

	rc.confState = snapshotToSave.Metadata.ConfState
	rc.snapshotIndex = snapshotToSave.Metadata.Index
	rc.appliedIndex = snapshotToSave.Metadata.Index*/
}

var snapshotCatchUpEntriesN uint64 = 10000

// Todo
func (rc *raftNode) maybeTriggerSnapshot() {
	/*	if rc.appliedIndex-rc.snapshotIndex <= rc.snapCount {
			return
		}

		log.Printf("start snapshot [applied index: %d | last snapshot index: %d]", rc.appliedIndex, rc.snapshotIndex)
		data, err := rc.getSnapshot()
		if err != nil {
			log.Panic(err)
		}
		snap, err := rc.raftStorage.CreateSnapshot(rc.appliedIndex, &rc.confState, data)
		if err != nil {
			panic(err)
		}
		if err := rc.saveSnap(snap); err != nil {
			panic(err)
		}

		compactIndex := uint64(1)
		if rc.appliedIndex > snapshotCatchUpEntriesN {
			compactIndex = rc.appliedIndex - snapshotCatchUpEntriesN
		}
		if err := rc.raftStorage.Compact(compactIndex); err != nil {
			panic(err)
		}

		log.Printf("compacted log at index %d", compactIndex)
		rc.snapshotIndex = rc.appliedIndex*/
}
