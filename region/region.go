package region

type Options struct {
	SeqId int
}

func DefaultRegionOptions() Options {
	return Options{}
}

func Open(opts Options) (region Region, err error) {
	// Load metadata
	// open the region
	return nil, nil
}

// Region needs to be created with a config
func Create(opts Options) (region Region, err error) {
	// create its metadata, memtable, sstable, and vlog
	region = &regionV1{
		seqId: opts.SeqId,
	}

	return region, nil
}

type Region interface {
	Close()
	Drop()
	Archive()
	Mutate()
	Get()
	Merge()
}
