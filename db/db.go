package db

type CreateConfig struct {
	// data path
}

func Create(config CreateConfig) (db *DB, err error) {
	return
}

// DB needs to be opened from a data path
func Open(path string) (db *DB, err error) {

	return nil, nil
}

type DB interface {
	Init()
	Close()
	ListTables()
	AddTable()
	DropTable()
	ArchiveTable()
}
