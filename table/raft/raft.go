package raft

import (
	"fmt"
	"github.com/coreos/etcd/raft/raftpb"
	"github.com/coreos/etcd/rafthttp"
)

type RNodeOptions struct {
	CurrentNodeId int
	Peers         []string
	Join          bool
	WalDir        string
}

type RNode struct {
	opts             RNodeOptions
	mutationStream   <-chan []byte
	commitStream     chan []byte
	confChangeStream <-chan raftpb.ConfChange

	transport          *rafthttp.Transport
	stopMutationStream chan struct{} // signals proposal channel closed
	raftTransportError chan struct{} // signals http server to shutdown
	raftDoneStream     chan struct{} // signals http server shutdown complete
}

func (rn *RNode) startRaft() {
	fmt.Println("Starting raft ......")

}

func NewRaftNode(opts RNodeOptions, mutationStream <-chan []byte, confChangeStream <-chan raftpb.ConfChange) *RNode {

	node := &RNode{
		opts:               opts,
		mutationStream:     mutationStream,
		commitStream:       make(chan []byte),
		confChangeStream:   confChangeStream,
		stopMutationStream: make(chan struct{}),
		raftTransportError: make(chan struct{}),
		raftDoneStream:     make(chan struct{}),
	}

	go node.startRaft()
	return node
}
