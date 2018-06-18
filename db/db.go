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
	return nil, nil
}

func OpenForDebug(id int, cluster string) (db *Handle, err error) {
	//log.Info("Opening data handle at path ")
	handle := &Handle{
		tables: make(map[string]table.Table),
	}
	handle.tables["t1"] = table.NewKvTable(id, cluster)

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
