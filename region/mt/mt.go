package mt

import (
	"github.com/dailyhunt/airdb/operation"
	"github.com/dailyhunt/airdb/region/airdb_skl"
)

const MB = 1024 * 1024

type MemtableOptions struct {
	MemtableFlushSizeInBytes uint64
}

func DefaultOptions() *MemtableOptions {
	return &MemtableOptions{
		MemtableFlushSizeInBytes: 16 * MB,
	}
}

type Memtable interface {
	Append(put *operation.Put) error
	SizeInBytes() uint64
	GetOptions() *MemtableOptions
	Len() int
	GetSeqNo() uint32
	SetSeqNo(seq uint32)
	Front() *airdb_skl.Element
}

func NewMemtable() Memtable {
	options := DefaultOptions()
	options.MemtableFlushSizeInBytes = 60
	return &sklMemtable{
		skipList: airdb_skl.NewWithMaxLevel(10),
		options:  options,
		seqNo:    1,
	}
}
