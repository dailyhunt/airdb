package db

type Config struct {
}

// DB needs to be opened with a config
func Open(config Config) (db *DB, err error) {
	return nil, nil
}

type DB interface {
	Close()
	AddTable()
	DropTable()
	ArchiveTable()
}
