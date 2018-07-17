package mt

import (
	skl "github.com/andy-kimball/arenaskl"
	"github.com/dailyhunt/airdb/utils/commonUtils"
)

var defaultArenaSize = commonUtils.MbToBytes(64)

type Memtable interface {
	Put(keyWithTs []byte, value []byte) error
	Iterator() *skl.Iterator
}

// Todo: Use options from Config file
func NewMemtable() *memtable {
	arena := skl.NewArena(uint32(defaultArenaSize))
	return &memtable{
		list: skl.NewSkiplist(arena),
	}
}

type memtable struct {
	list *skl.Skiplist
}

func (m *memtable) Iterator() *skl.Iterator {
	return m.getItr()
}

func (m *memtable) getItr() *skl.Iterator {
	var it skl.Iterator
	it.Init(m.list)
	return &it
}

func (m *memtable) Put(keyWithTs []byte, value []byte) error {
	// Todo : Should be done in constructor function or it fine to do here
	it := m.getItr()
	return it.Add(keyWithTs, value, 0)
}
