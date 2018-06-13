package region

type Config struct {
}

func Open(name string) (region *Region, err error) {
	// Load metadata
	// open the region
	return nil, nil
}

// Region needs to be created with a config
func Create(config Config) (region *Region, err error) {
	// create its metadata, memtable, sstable, and vlog
	return nil, nil
}

type Region interface {
	Close()
	Drop()
	Archive()
	Put()
	Get()
	Merge()
	Add()
	Decay()
}
