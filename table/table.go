package table

import (
	"github.com/dailyhunt/airdb/airaft"
	"github.com/dailyhunt/airdb/operation"
	"github.com/dailyhunt/airdb/region"
	"github.com/dailyhunt/airdb/region/mt"
	"github.com/dailyhunt/airdb/region/vlog"
	logger "github.com/sirupsen/logrus"
	"strings"
)

type Config struct {
}

func Open(name string) (table *Table, err error) {
	// Load metadata
	// open all regions
	return nil, nil
}

// Table needs to be created with a config
func Create(config Config) (table *Table, err error) {
	return nil, nil
}

func NewKvTable(id int, cluster string) (table Table) {

	logger.Info("Added dummy table with local fs region : Table Name t1")

	// Todo(sohan) : Need to move to Region class as new method
	localFsRegion := &region.LocalFs{}
	localFsRegion.ID = 1
	localFsRegion.IsLeader = true
	localFsRegion.MemTable = mt.New()
	localFsRegion.Raft = airaft.NewRaft(id, strings.Split(cluster, ","))
	localFsRegion.VLog = vlog.New()

	t := &KvTable{
		Region: localFsRegion,
	}

	return t

}

type Table interface {
	Close()
	Drop()
	Archive()
	Put(put *operation.Put) error
	Get()
	Merge()
	Add()
	Decay()
}
