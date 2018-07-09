package table

import (
	"context"
	"github.com/dailyhunt/airdb/region"
	"github.com/dailyhunt/airdb/utils/fileUtils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	MaxRegionsPerTable = 1
)

var emptyTablePathError = errors.New("table path is empty.")
var tableDirDoesNotExists = errors.New("table dir does not exist.")

type Options struct {
	Name           string
	Path           string
	SstDir         string
	WalDir         string
	MaxRegions     int
	RegionSeqId    int64
	CurrentRegions int
}

type Manifest struct {
	Version       int
	ActiveRegions []int64
	Options       Options
}

func DefaultTableOptions() Options {
	return Options{
		MaxRegions:     MaxRegionsPerTable,
		RegionSeqId:    0,
		CurrentRegions: 0,
	}
}

type Table interface {
	Open(option *Options)
	Close()
	Drop()
	Archive()
	// Insert / Update / Increment operation ( decay also - later)
	Mutate(ctx context.Context, data []byte) error
	Get()
	Merge()
	Add()
	Decay()
	Name() string
}

func NewTable(opts Options) (Table, error) {
	t := &tableImpl{
		name:    opts.Name,
		regions: make(map[int64]region.Region),
	}

	// Read from manifest

	rgOpts := region.DefaultRegionOptions()
	rgOpts.SeqId = 1

	r, err := region.Create(rgOpts)
	if err != nil {
		return nil, err
	}

	t.regions[1] = r

	return t, nil
}

func OpenTable(path string) (Table, error) {
	if path == "" {
		return nil, emptyTablePathError
	}

	exists, err := fileUtils.Exists(path)
	if !exists || err != nil {
		return nil, tableDirDoesNotExists
	}

	log.Infof("opening table at path %s ", path)
	m, err := readManifestFile(path)
	if err != nil {
		return nil, err
	}

	t := NewTableImpl(m)

	// Todo : Open region properly from manifest
	t.regions = make(map[int64]region.Region)
	rg, _ := region.Create(region.DefaultRegionOptions())
	t.regions[0] = rg

	return t, nil

}
