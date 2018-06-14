package db

import (
	"github.com/dailyhunt/airdb/table"
)

type CreateConfig struct {
	// data path
}

func Create(config CreateConfig) (db *Handle, err error) {
	return
}

// DB needs to be opened from a data path
func Open(path string) (db *Handle, err error) {
	handle := &Handle{
		tables: make(map[string]*table.KvTable),
	}

	handle.tables["t1"] = &table.KvTable{}

	return handle, nil
}

type DB interface {
	Init()
	Close()
	ListTables()
	AddTable()
	DropTable()
	ArchiveTable()
}
