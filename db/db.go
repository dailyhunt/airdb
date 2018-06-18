package db

import "github.com/dailyhunt/airdb/table"

type CreateConfig struct {
	// data path
	path string
}

//type InitConfig struct {
//}

func Create(config CreateConfig) (db *DB, err error) {
	return
}

// DB needs to be opened from a data path
func Open(path string) (db *DB, err error) {

	return nil, nil
}

type DB interface {
	//Init(config InitConfig) (err error)
	Close() (err error)
	ListTables(inactive bool) (tables []table.Table, err error)
	GetTable(name string) (table table.Table, err error)
	CreateTable(config table.CreateConfig) (table table.Table, err error)
	DropTable(name string) (err error)
	ArchiveTable(name string) (err error)
}
