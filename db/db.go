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
		tables: make(map[string]table.Table),
	}

	handle.tables["t1"] = &table.KvTable{}

	return handle, nil
}

type DB interface {
	Init()
	Close()
	GetTable(name string) (table.Table, error)
	ListTables()
	AddTable()
	DropTable()
	ArchiveTable()
}
