package airaft

import (
	"github.com/coreos/etcd/etcdserver/api/rafthttp"
	"github.com/coreos/etcd/raft"
	"github.com/coreos/etcd/raft/raftpb"
	"github.com/coreos/etcd/wal"
)

type raftNode struct {
	// Basic Raft info
	ID    int      // id for raft session
	join  bool     // flag , if node is joining existing cluster
	peers []string // url of raft peers

	// TODO: (shailendra)channel should be of other type instead of string
	// TODO: Type still needed to decide
	proposeCh        <-chan string            // proposed key value
	raftConfChangeCh <-chan raftpb.ConfChange // proposed raft/region config change
	commitLogCh      chan<- *string           // entries committed to log/wal for (key,value)
	errorCh          chan<- error             // errors from raft session

	// wal
	walDir string
	wal    *wal.WAL

	// snapshot
	snapDir string

	// raft
	raftState     raftpb.ConfState
	raftNode      raft.Node
	raftStorage   *raft.MemoryStorage
	raftTransport *rafthttp.Transport

	// Stop/Done channels
	// TODO: (sohan)
	stopCh     chan struct{}
	httpStopCh chan struct{}
	httpDoneCh chan struct{}
}

func newRaftNode() {

}

func (rc *raftNode) startRaft() {

}

func (rc *raftNode) stopRaft() {

}

// Find out entries to apply/publish
// Ignore old entries from last index
func (rc *raftNode) entriesToApply() {

}

// Write committed logs to commit Channel
// Return if all entries could be published
func (rc *raftNode) publishEntries() {

}

func (rc *raftNode) openWal() {

}

func (rc *raftNode) replayWAL() {

}
