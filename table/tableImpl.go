package table

import (
	"context"
	"github.com/coreos/etcd/raft/raftpb"
	"github.com/coreos/etcd/wal"
	"github.com/dailyhunt/airdb/mutation"
	r "github.com/dailyhunt/airdb/region"
	raft "github.com/dailyhunt/airdb/table/raft"
	log "github.com/sirupsen/logrus"
	"os"
)

type tableImpl struct {
	name             string
	regions          map[int64]r.Region
	manifest         *Manifest
	opts             Options
	mutationStream   <-chan []byte
	confChangeStream <-chan raftpb.ConfChange
	wal              *wal.WAL
	raft             *raft.RNode
}

func NewTableImpl(m *Manifest) *tableImpl {
	t := &tableImpl{
		name:     m.Options.Name,
		manifest: m,
		opts:     m.Options,
	}

	t.initializeWal()
	t.mutationStream = make(chan []byte)
	t.confChangeStream = make(chan raftpb.ConfChange)

	rNodeOptions := raft.RNodeOptions{
		CurrentNodeId: 0,
		Peers:         []string{},
		Join:          false,
		WalDir:        t.opts.WalDir,
	}

	t.raft = raft.NewRaftNode(rNodeOptions, t.mutationStream, t.confChangeStream)

	return t
}

func (t *tableImpl) Name() string {
	return t.name
}

func (t *tableImpl) Open(option *Options) {
	panic("implement me")
}

func (t *tableImpl) Close() {
	t.FlushManifest()
}

func (t *tableImpl) Drop() {
	panic("implement me")
}

func (t *tableImpl) Archive() {
	panic("implement me")
}

func (t *tableImpl) Mutate(ctx context.Context, data []byte) error {
	m, _ := mutation.Decode(data)
	log.Infof("proposing mutation to region ... %s ", string(m.Key))
	return nil
}

func (t *tableImpl) Get() {
	panic("implement me")
}

func (t *tableImpl) Merge() {
	panic("implement me")
}

func (t *tableImpl) Add() {
	panic("implement me")
}

func (t *tableImpl) Decay() {
	panic("implement me")
}
func (t *tableImpl) FlushManifest() {

}

func (t *tableImpl) initializeWal() {
	walDir := t.opts.WalDir

	// Create wal if not exist
	if !wal.Exist(walDir) {
		if err := os.Mkdir(walDir, 0750); err != nil {
			log.Fatalf("cannot create wal dir at %s for table %s due to (%v)", walDir, t.Name(), err)
		}

		w, err := wal.Create(walDir, nil)
		if err != nil {
			log.Fatalf("cannot create wal file at %s for table %s due to (%v)", walDir, t.Name(), err)
		}
		w.Close()
	}

}
