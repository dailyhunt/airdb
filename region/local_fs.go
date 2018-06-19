package region

import (
	//	"bytes"
	//	"encoding/gob"
	"github.com/coreos/etcd/etcdserver/api/snap"
	"github.com/dailyhunt/airdb/operation"
	log "github.com/sirupsen/logrus"

	"encoding/json"
)

type LocalFs struct {
	proposeC chan<- string // channel for proposing updates
	//mu          sync.RWMutex
	kvStore     map[string]string // current committed key-value pairs
	snapshotter *snap.Snapshotter
	Replica
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
			log.Printf("loading snapshot at term %d and index %d", snapshot.Metadata.Term, snapshot.Metadata.Index)
			if err := fs.recoverFromSnapshot(snapshot.Data); err != nil {
				log.Error("not able to recover snapshot")
				log.Panic(err)
			}
			continue
		}

		var dataKv map[string]string
		err := json.Unmarshal([]byte(*data), &dataKv)
		log.Debug("Reading commits - kv size ", len(fs.kvStore))
		if err != nil {
			log.Fatalf("airdb: could not decode message (%v)", err)
		}
		//.mu.Lock()
		fs.kvStore[dataKv["K"]] = dataKv["V"]

		log.Info("Kv Store with size  ", len(fs.kvStore))

		//s.mu.Unlock()
	}
	if err, ok := <-errorC; ok {
		log.Fatal(err)
		close(fs.proposeC)
		//close(errorC) // todo
	}
}

func (fs *LocalFs) GetSnapshot() ([]byte, error) {
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

func NewLocalFs(snapshotter *snap.Snapshotter, proposeChan chan<- string, commitChan <-chan *string, errorChan <-chan error) *LocalFs {

	region := &LocalFs{
		proposeC:    proposeChan,
		kvStore:     make(map[string]string),
		snapshotter: snapshotter,
	}

	//region.proposeC <- "hello"

	region.readCommits(commitChan, errorChan)
	go region.readCommits(commitChan, errorChan)

	return region

}
