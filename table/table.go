package table

import (
	"context"
	"github.com/dailyhunt/airdb/region"
)

type Options struct {
	Name        string
	Path        string
	SstDir      string
	WalDir      string
	MaxRegions  int
	RegionSeqId int
}

func DefaultTableOptions() Options {
	return Options{}
}

type Table interface {
	Open(option *Options)
	Close()
	Drop()
	Archive()
	Put(ctx context.Context, data []byte) error
	Get()
	Merge()
	Add()
	Decay()
	Name() string
}

func NewTable(opts Options) (Table, error) {
	t := &tableImpl{
		name:    opts.Name,
		regions: make(map[int]region.Region),
	}

	rgOpts := region.DefaultRegionOptions()
	rgOpts.SeqId = 1

	region, err := region.Create(rgOpts)
	if err != nil {
		return nil, err
	}

	t.regions[1] = region

	return t, nil
}
