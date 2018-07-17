package region

import (
	"context"
	"github.com/andy-kimball/arenaskl"
	"github.com/coreos/etcd/raft/raftpb"
	"github.com/dailyhunt/airdb/mutation"
	"github.com/dailyhunt/airdb/region/mt"
	"github.com/dailyhunt/airdb/region/raft"
	log "github.com/sirupsen/logrus"
	"sync"
)

type regionV1 struct {
	seqId            int
	mutationStream   chan<- []byte
	confChangeStream <-chan raftpb.ConfChange
	raft             *raft.RNode
	mt               mt.Memtable
	imt              mt.Memtable // immutable memtable
	mu               sync.Mutex
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

	for commitEntry := range commitStream {
		data := commitEntry.Data

		if data == nil {
			log.Info("TODO: Refer etcd example  ... Got nil commitEntry ...done with replaying WAl ..")
			return
		}

		keyWithTs := mutation.KeyWithTs(data)
		err := r.mt.Put(keyWithTs, data)
		if err != nil {
			r.mayBeFlushMemtable(commitEntry, err)
			r.mt.Put(keyWithTs, data)
		}

		// Todo : Read and forward to respective region
		/*m, _ := mutation.Decode(commitEntry)
		fmt.Println(m.String())*/
	}

	errorStream := r.raft.GetErrorStream()
	if err, ok := <-errorStream; ok {
		log.Fatal(err)
	}

}
func (r *regionV1) mayBeFlushMemtable(entry raftpb.Entry, e error) {
	if e == arenaskl.ErrArenaFull {
		// Flush Memtable
		r.mu.Lock()
		r.swapAndCreateNewMemTable()
		go r.flushIMemtable(entry.Index)
		r.mu.Unlock()
		return

	}

	if e == arenaskl.ErrRecordExists {
		//TODO : Handle or not
		return
	}

}
func (r *regionV1) swapAndCreateNewMemTable() {
	m := r.mt
	r.imt = m
	r.mt = mt.NewMemtable()
	// Todo::
	// Increment sequence num etc

}
func (r *regionV1) flushIMemtable(indexForSnapshot uint64) {
	if r.imt == nil {
		log.Fatal("flushIMemtable() is called when fs.imemtable == nil")
	}
	log.Info("flushing memtable")
	lastSuccessfulProcessedIndex := indexForSnapshot - 1
	defer r.raft.MaybeTriggerSnapshot(lastSuccessfulProcessedIndex)

	m := r.imt
	itr := m.Iterator()
	for itr.SeekToFirst(); itr.Valid(); itr.Next() {
		key := mutation.ParseKey(itr.Key())
		mut, _ := mutation.Decode(itr.Value())
		log.Debugf("Key %s :: Value %s ", string(key), mut.String())
	}

	r.imt = nil

}
