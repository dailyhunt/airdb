package region

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/raft/raftpb"
	"github.com/dailyhunt/airdb/mutation"
	"github.com/dailyhunt/airdb/region/raft"
	log "github.com/sirupsen/logrus"
)

type regionV1 struct {
	seqId            int
	mutationStream   chan<- []byte
	confChangeStream <-chan raftpb.ConfChange
	raft             *raft.RNode
}

func (r *regionV1) GetRegionId() int {
	return r.seqId
}

func (r *regionV1) Close() {
	panic("implement me")
}

func (r *regionV1) Drop() {
	panic("implement me")
}

func (r *regionV1) Archive() {
	panic("implement me")
}

func (r *regionV1) Mutate(ctx context.Context, data []byte) {
	r.mutationStream <- data
}

func (r *regionV1) Get() {
	panic("implement me")
}

func (r *regionV1) Merge() {
	panic("implement me")
}

func (r *regionV1) readCommitStream() {
	commitStream := r.raft.GetCommitStream()

	for commit := range commitStream {
		if commit == nil {
			log.Info("TODO: Refer etcd example  ... Got nil commit ...done with replaying WAl ..")
			return
		}

		// Todo : Read and forqard to respective region
		fmt.Println(mutation.Decode(commit))
	}

	errorStream := r.raft.GetErrorStream()
	if err, ok := <-errorStream; ok {
		log.Fatal(err)
	}

}
