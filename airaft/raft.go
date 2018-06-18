package airaft

import (
	"github.com/coreos/etcd/etcdserver/api/rafthttp"
	"github.com/coreos/etcd/raft"
	"github.com/coreos/etcd/raft/raftpb"
	"github.com/coreos/etcd/wal"

	"context"
	"encoding/json"
	stats "github.com/coreos/etcd/etcdserver/api/v2stats"
	"github.com/coreos/etcd/pkg/types"
	"github.com/dailyhunt/airdb/operation"
	logger "github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type RaftNode struct {
	// Basic Raft info
	ID    int      // id for raft session
	join  bool     // flag , if node is joining existing cluster
	peers []string // url of raft peers

	// TODO: (shailendra)channel should be of other type instead of string
	// TODO: Type still needed to decide
	proposeCh        <-chan operation.Op      // proposed key value
	raftConfChangeCh <-chan raftpb.ConfChange // proposed raft/region config change
	commitLogCh      chan<- operation.Op      // entries committed to log/wal for (key,value)
	errorCh          chan<- error             // errors from raft session

	// wal
	walDir string
	wal    *wal.WAL

	// snapshot
	snapDir string

	// raft
	raftState        raftpb.ConfState
	raftNode         raft.Node
	raftStorage      *raft.MemoryStorage
	raftTransport    *rafthttp.Transport
	lastAppliedIndex uint64
	lastIndex        uint64 // index of log at start

	// Explicit cancellation of channels
	stopCh     chan struct{}
	httpStopCh chan struct{}
	httpDoneCh chan struct{}
}

// Todo(sohan): Error handling
func NewRaft(id int, peers []string) *RaftNode {

	rn := &RaftNode{
		ID:               id,
		peers:            peers,
		proposeCh:        make(chan operation.Op),
		raftConfChangeCh: make(chan raftpb.ConfChange),
		commitLogCh:      make(chan operation.Op),
		errorCh:          make(chan error),

		stopCh:     make(chan struct{}),
		httpStopCh: make(chan struct{}),
		httpDoneCh: make(chan struct{}),
	}

	go rn.startRaft()
	go rn.serveChannels()

	return rn

}

func (rc *RaftNode) startRaft() {

	// Todo(sohan): Snapshot
	// Todo(sohan): WAL

	// Todo(sohan): Raft Config for peers
	rcPeers := make([]raft.Peer, len(rc.peers))
	for i := range rc.peers {
		rcPeers[i] = raft.Peer{ID: uint64(i + 1)} // Todo - doc
	}

	// TODO(sohan) - set read only replica using learners array
	logger.Info("Creating Raft Config for ", rc.ID)

	// Todo(Sohan) - take config from Config file
	rConfig := &raft.Config{
		ID:              uint64(rc.ID),
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         rc.raftStorage,
		MaxSizePerMsg:   1024 * 1024,
		MaxInflightMsgs: 256,
	}

	// Todo - Start vs restart node
	startPeers := rcPeers
	if rc.join {
		startPeers = nil
	}
	rc.raftNode = raft.StartNode(rConfig, startPeers)

	// Todo(sohan): Raft Transport
	rc.raftTransport = &rafthttp.Transport{
		Logger:      zap.NewExample(),
		ID:          types.ID(rc.ID),
		ClusterID:   0x1000,
		Raft:        rc,
		ServerStats: stats.NewServerStats("", ""),
		LeaderStats: stats.NewLeaderStats(strconv.Itoa(rc.ID)),
		ErrorC:      make(chan error),
	}

	logger.WithFields(logger.Fields{
		"clusterId":   rc.raftTransport.ClusterID,
		"transportId": rc.raftTransport.ID,
	}).Info("Starting Raft Transport")

	rc.raftTransport.Start()

	for i := range rc.peers {
		if i+1 != rc.ID {
			logger.WithFields(logger.Fields{
				"id":        rc.raftTransport.ID,
				"peerId":    rc.peers[i],
				"clusterId": rc.raftTransport.ClusterID,
			}).Info("Adding Peer to Transport")

			rc.raftTransport.AddPeer(types.ID(i+1), []string{rc.peers[i]})
		}
	}

	go rc.serveRaft()
	go rc.serveChannels()

}

// Todo: Doc
func (rc *RaftNode) serveRaft() {
	logger.Info("Starting Serving Raft transport", rc.ID)
	transportAddress := rc.peers[rc.ID-1]
	url, err := url.Parse(transportAddress)
	if err != nil {
		logger.Fatalf("Failing parsing URL for raft transport (%v)", err)
	}
	logger.Info("Raft Transport Address ", url)
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

func (rc *RaftNode) serveChannels() {
	// Todo : Wal and snapshot related

	// Todo(sohan): Take from config
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	go rc.proposeToRaft()

	// Event loop on raft state machine updates
	for {
		select {
		case <-ticker.C:
			rc.raftNode.Tick()

		// store raft entries to wal, then publish over commit channel
		case rd := <-rc.raftNode.Ready():
			// Todo:
			//rc.wal.Save(rd.HardState, rd.Entries)
			// Todo: Save to WAL
			// Todo: Snapshot stuff
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

}

func (rc *RaftNode) proposeToRaft() {
	var configChangeCount uint64 = 0 // for config change id

	for rc.proposeCh != nil && rc.raftConfChangeCh != nil {
		select {
		case proposed, ok := <-rc.proposeCh:
			if !ok {
				rc.proposeCh = nil
			} else {
				// Todo(sohan)- switch based on type
				op := proposed.Name()
				switch op {
				case "PUT":
					logger.Debug("Processing Put in proposeRaft ", proposed.String())
					rc.raftNode.Propose(context.TODO(), []byte(proposed.String()))
				default:
					// TODO
					logger.Warn("Invalid operation")
				}

			}

		case configChange, ok := <-rc.raftConfChangeCh:
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

func (rc *RaftNode) stopRaft() {
	rc.stopHTTP()
	close(rc.commitLogCh)
	close(rc.errorCh)
	rc.raftNode.Stop()
}

func (rc *RaftNode) writeError(err error) {
	rc.stopHTTP()
	close(rc.commitLogCh)
	rc.errorCh <- err
	close(rc.errorCh)
	rc.raftNode.Stop()
}

func (rc *RaftNode) stopHTTP() {
	rc.raftTransport.Stop()
	close(rc.httpStopCh)
	<-rc.httpDoneCh
}

// Find out entries to apply/publish
// Ignore old entries from last index
func (rc *RaftNode) entriesToApply(inEntries []raftpb.Entry) (outEntries []raftpb.Entry) {
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
func (rc *RaftNode) publishEntries(entries []raftpb.Entry) bool {
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
			var put operation.Put
			json.Unmarshal(entry.Data, &put)

			select {
			case rc.commitLogCh <- &put: // todo (pointer allocation)
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
					rc.raftTransport.AddPeer(nodeId, us)
				}
			case raftpb.ConfChangeRemoveNode:
				if cc.NodeID == uint64(rc.ID) {
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
			select {
			case rc.commitLogCh <- nil:
			case <-rc.stopCh:
				return false
			}
		}
	} // for

	return true
}

func (rc *RaftNode) openWal() {

}

func (rc *RaftNode) replayWAL() {

}

func (rc *RaftNode) Process(ctx context.Context, m raftpb.Message) error {
	return rc.raftNode.Step(ctx, m)
}

func (rc *RaftNode) IsIDRemoved(id uint64) bool {
	panic("implement me")
}

func (rc *RaftNode) ReportUnreachable(id uint64) {
	panic("implement me")
}

func (rc *RaftNode) ReportSnapshot(id uint64, status raft.SnapshotStatus) {
	panic("implement me")
}
