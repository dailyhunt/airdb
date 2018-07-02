package table

import (
	"github.com/coreos/etcd/raft/raftpb"
	"github.com/dailyhunt/airdb/airaft"
	"github.com/dailyhunt/airdb/operation"
	"github.com/dailyhunt/airdb/region"
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

func NewKvTable(id int, cluster string, join bool) (table Table) {

	logger.Debug("Added dummy table with local fs region : Table Name t1")

	proposeC := make(chan string)
	confChangeC := make(chan raftpb.ConfChange)

	var localFsRegion *region.LocalFs

	//localFsRegion := &region.LocalFs{}

	// raft provides a commit stream for the proposals from the http api
	getSnapshot := func() ([]byte, error) { return localFsRegion.GetSnapshot() }
	commitC, errorC, snapshotterReady := airaft.NewRaftNode(id, strings.Split(cluster, ","), join, getSnapshot, proposeC, confChangeC)

	localFsRegion = region.NewLocalFs(<-snapshotterReady, proposeC, commitC, errorC, confChangeC)

	t := &KvTable{
		Region: localFsRegion,
	}

	//localFsRegion.ReadCommits(commitC, errorC)
	//go localFsRegion.ReadCommits(commitC, errorC)

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
	AddRegionPeer(nodeId int64, url []byte) error
}
