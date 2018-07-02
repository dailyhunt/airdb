package mt

import (
	"github.com/dailyhunt/airdb/operation"
	list "github.com/dailyhunt/airdb/region/airdb_skl"
	"sync/atomic"
)

type sklMemtable struct {
	totalSize uint64
	seqNo     uint32
	skipList  *list.SkipList
	options   *MemtableOptions
}

func (skl *sklMemtable) Front() *list.Element {
	return skl.skipList.Front()
}

func (skl *sklMemtable) SetSeqNo(seq uint32) {
	skl.seqNo = seq
}

func (skl *sklMemtable) GetSeqNo() uint32 {
	return skl.seqNo
}

func (skl *sklMemtable) GetOptions() *MemtableOptions {
	return skl.options
}

func (skl *sklMemtable) SizeInBytes() uint64 {
	return skl.totalSize
}

// Upsert
func (skl *sklMemtable) Append(put *operation.Put) error {
	skl.skipList.Set(put.K, put)
	atomic.AddUint64(&skl.totalSize, put.Size())
	return nil
}

func (skl *sklMemtable) Len() int {
	return skl.skipList.Length
}
