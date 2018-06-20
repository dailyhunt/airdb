package region

import (
	"github.com/dailyhunt/airdb/operation"
	"github.com/dailyhunt/airdb/region/mt"
	"github.com/dailyhunt/airdb/region/vlog"
)

type Config struct {
}

func Open(name string) (region *Region, err error) {
	// Load metadata
	// open the region
	return nil, nil
}

// Region needs to be created with a config
func Create(config Config) (region *Region, err error) {
	// create its metadata, memtable, sstable, and vlog
	return nil, nil
}

type Region interface {
	Close()
	Drop()
	Archive()
	Put(put *operation.Put) error
	Get()
	Merge()
	Add()
	Decay()
	AddPeer(nodeId int64, url []byte)
	//ReadCommits()
}

type Replica struct {
	ID       uint64
	IsLeader bool
	MemTable *mt.Memtable
	VLog     *vlog.Vlog
	// Ids of peers for this replica across cluster
	Peers []uint64
}
