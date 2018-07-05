package table

import (
	"context"
)

type Options struct {
	path       string
	sstDir     string
	walDir     string
	maxRegions int
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
}

func NewTable() {

}
