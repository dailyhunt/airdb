package table

import (
	"context"
)

type CreateConfig struct {
}

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

type Table interface {
	Close()
	Drop()
	Archive()
	Put(ctx context.Context, data []byte) error
	Get()
	Merge()
	Add()
	Decay()
}
