package region

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/raft/raftpb"
	"github.com/dailyhunt/airdb/region/raft"
	"github.com/spf13/viper"
)

type Options struct {
	SeqId  int
	WalDir string
}

func DefaultRegionOptions() Options {
	return Options{
		SeqId: 999,
	}
}

func Open(opts Options) (region Region, err error) {
	// Load metadata
	// open the region
	return nil, nil
}

// Region needs to be created with a config
func Create(opts Options) (region Region, err error) {
	// create its metadata, memtable, sstable, nd vlog

	mutationStream := make(chan []byte)

	rg := &regionV1{
		seqId:            opts.SeqId,
		mutationStream:   mutationStream,
		confChangeStream: make(chan raftpb.ConfChange),
	}

	rNodeOptions := raft.RNodeOptions{
		CurrentNodeId:  1,
		Peers:          []string{},
		Join:           false,
		WalDir:         opts.WalDir,
		RaftPort:       viper.GetInt("raft.port"),
		TickerInMillis: int64(viper.GetInt("raft.tickerInMill")),
	}

	// TODO :: Id will region id and should be across database
	rNodeOptions.Peers = append(rNodeOptions.Peers, fmt.Sprintf("http://localhost:%d", rNodeOptions.RaftPort))

	rg.raft = raft.NewRaftNode(rNodeOptions, mutationStream, rg.confChangeStream)
	rg.readCommitStream()
	go rg.readCommitStream()

	return rg, nil
}

type Region interface {
	GetRegionId() int
	Close()
	Drop()
	Archive()
	Mutate(ctx context.Context, data []byte)
	Get()
	Merge()
}
