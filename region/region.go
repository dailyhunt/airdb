package region

import (
	"github.com/dailyhunt/airdb/airaft"
	"github.com/dailyhunt/airdb/region/mt"
	"github.com/dailyhunt/airdb/region/vlog"
	"github.com/dailyhunt/airdb/table"
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
	Put(put *table.Put) error
	Get()
	Merge()
	Add()
	Decay()
}

type Replica struct {
	ID       uint64
	isLeader bool
	raft     airaft.RaftNode
	memTable mt.Memtable
	vLog     vlog.Vlog
	// Ids of peers for this replica across cluster
	peers []uint64
}
