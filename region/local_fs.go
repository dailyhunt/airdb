package region

import (
	//	"bytes"
	//	"encoding/gob"
	"github.com/coreos/etcd/etcdserver/api/snap"
	"github.com/dailyhunt/airdb/operation"
	log "github.com/sirupsen/logrus"

	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/raft/raftpb"
	"github.com/dailyhunt/airdb/region/mt"
	"github.com/dailyhunt/airdb/region/sst"
	"github.com/dailyhunt/airdb/region/vlog"
	"os"
	"sync"
)

type LocalFs struct {
	proposeC    chan<- string // channel for proposing updates
	confChangeC chan<- raftpb.ConfChange
	kvStore     map[string]string // current committed key-value pairs
	snapshotter *snap.Snapshotter
	mu          sync.Mutex
	memtable    mt.Memtable
	imemtable   mt.Memtable // Immutable memtable
	Replica
}

func (fs *LocalFs) AddPeer(nodeId int64, url []byte) {
	cc := raftpb.ConfChange{
		Type:    raftpb.ConfChangeAddNode,
		NodeID:  uint64(nodeId),
		Context: url,
	}
	fs.confChangeC <- cc
}

func (fs *LocalFs) Close() {
	panic("implement me")
}

func (fs *LocalFs) Drop() {
	panic("implement me")
}

func (fs *LocalFs) Archive() {
	panic("implement me")
}

func (fs *LocalFs) Put(put *operation.Put) error {
	marshal, _ := json.Marshal(put)
	fs.proposeC <- string(marshal)
	return nil
}

func (fs *LocalFs) Get() {
	panic("implement me")
}

func (fs *LocalFs) Merge() {
	panic("implement me")
}

func (fs *LocalFs) Add() {
	panic("implement me")
}

func (fs *LocalFs) Decay() {
	panic("implement me")
}

func (fs *LocalFs) MayBeTriggerFlush() {
	if fs.shouldFlush() {
		fs.mu.Lock()
		if fs.shouldFlush() {
			fs.swapAndCreateNewMemTable()
			go fs.flushIMemtable()
		}
		fs.mu.Unlock()
	}
}

func (fs *LocalFs) shouldFlush() bool {
	maxFlushSize := fs.memtable.GetOptions().MemtableFlushSizeInBytes
	return fs.memtable.SizeInBytes() > maxFlushSize && fs.imemtable == nil
}

func (fs *LocalFs) swapAndCreateNewMemTable() {
	m := fs.memtable
	fs.imemtable = m
	fs.memtable = mt.NewMemtable()
	fs.memtable.SetSeqNo(m.GetSeqNo() + 1)
}

func (fs *LocalFs) readCommits(commitC <-chan *string, errorC <-chan error) {
	for data := range commitC {
		if data == nil {
			// done replaying log; new data incoming
			// OR signaled to load snapshot
			log.Debug("local_fs.go Nil data from commit chan")
			snapshot, err := fs.snapshotter.Load()
			if err == snap.ErrNoSnapshot {
				log.Debug("No snapshot : returning ")
				return
			}
			if err != nil {
				log.Panic(err)
			}
			log.Info("loading snapshot at term %d and index %d", snapshot.Metadata.Term, snapshot.Metadata.Index)
			if err := fs.recoverFromSnapshot(snapshot.Data); err != nil {
				log.Error("not able to recover snapshot")
				log.Panic(err)
			}
			continue
		}

		fs.MayBeTriggerFlush()

		// Todo(sohan) - want to actual deserialization/ unmarshelling ???
		// as raft only accepts data in byte[]

		var p operation.Put
		err := json.Unmarshal([]byte(*data), &p)

		/*		var dataKv map[string]string
				err := json.Unmarshal([]byte(*data), &dataKv)
				key := dataKv["K"]
				value := dataKv["V"]*/
		if err != nil {
			log.Fatalf("airdb: could not decode message (%v)", err)
		}
		/*//.mu.Lock()
		fs.kvStore[key] = value
		log.WithFields(log.Fields{
			"key":       key,
			"value":     value,
			"storeSize": len(fs.kvStore),
			"epoch":     utils.GetCurrentTime(),
		}).Info("Applying Commit to FSM ")*/

		fs.memtable.Append(&p)
		//log.Info("Kv Store with size  ", len(fs.kvStore))

		log.WithFields(log.Fields{
			"key":            p.K,
			"mtLen":          fs.memtable.Len(),
			"mtSize":         fs.memtable.SizeInBytes(),
			"mtSeq":          fs.memtable.GetSeqNo(),
			"immutableTable": fs.memtable.GetSeqNo(),
		}).Info("Inserted Element")

		//log.Info("Inserted ", p.K, " memtable length ", fs.memtable.Len(), " size in bytes ", fs.memtable.SizeInBytes())

		//s.mu.Unlock()

	}
	if err, ok := <-errorC; ok {
		log.Fatal(err)
		close(fs.proposeC)
		//close(errorC) // todo
	}
}

// TODO: this should not be here - not needed
func (fs *LocalFs) GetSnapshot() ([]byte, error) {
	log.Println("Starting creating snapshot")
	return json.Marshal(fs.kvStore)

}

func (fs *LocalFs) recoverFromSnapshot(Data []byte) interface{} {
	log.Debug("Recovering from snapshot")
	var store map[string]string
	if err := json.Unmarshal(Data, &store); err != nil {
		return err
	}
	//s.mu.Lock()
	fs.kvStore = store
	//s.mu.Unlock()
	return nil
}
func (fs *LocalFs) flushIMemtable() error {
	if fs.imemtable == nil {
		log.Fatal("flushIMemtable() is called when fs.imemtable == nil")
	}

	im := fs.imemtable
	path := "/home/sohanvir/softwares/go/src/github.com/dailyhunt/sst/001.sst"
	f, err := os.Open(path)
	vLog, err := os.OpenFile("/home/sohanvir/softwares/go/src/github.com/dailyhunt/sst/001.vlog",
		os.O_APPEND|os.O_WRONLY, 0666)

	if err != nil {
		panic(err)
	}

	builder := sst.NewSstBuilder()
	defer builder.Close()

	element := im.Front()

	for element != nil {
		key := element.Key()

		// Process Operations and combine then and write to files
		put := element.Value().(*operation.Put)

		// Encode and Decode keys

		element = element.Next()
	}

	sstBytesBuffer := builder.Finish()
	_, err = f.Write(sstBytesBuffer)
	fmt.Println("implement flushIMemtable ")

	return err
}

func NewLocalFs(snapshotter *snap.Snapshotter, proposeChan chan<- string, commitChan <-chan *string, errorChan <-chan error,
	confChange chan<- raftpb.ConfChange) *LocalFs {

	region := &LocalFs{
		proposeC:    proposeChan,
		confChangeC: confChange,
		kvStore:     make(map[string]string),
		snapshotter: snapshotter,
	}

	// Todo(sohan) : Need to move to Region class as new method
	region.ID = 1
	region.IsLeader = true
	region.memtable = mt.NewMemtable()
	//localFsRegion.Raft = airaft.NewRaft(id, strings.Split(cluster, ","))
	region.VLog = vlog.New()

	//region.proposeC <- "hello"

	region.readCommits(commitChan, errorChan)
	go region.readCommits(commitChan, errorChan)

	return region

}
