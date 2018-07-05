package db

import "github.com/dailyhunt/airdb/table"

type Options struct {
	// data path
	Path string
}

func DefaultDbOptions() Options {
	return Options{}
}

//type InitConfig struct {
//}

func Create(opts Options) (db DB, err error) {
	return
}

// DB needs to be opened from a data path
func Open(opts Options) (db DB, err error) {
	// Read manifest file and load table
	d := &Handle{}
	d.AddTestTable()
	return d, nil
}

type DB interface {
	//Init(config InitConfig) (err error)
	Close() (err error)
	ListTables(inactive bool) (tables []table.Table, err error)
	GetTable(name string) (table table.Table, err error)
	CreateTable(config table.Options) (table table.Table, err error)
	DropTable(name string) (err error)
	ArchiveTable(name string) (err error)
}
