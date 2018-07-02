package design

type Slice []byte

type Cell struct {
	key    Slice
	family Slice
	col    Slice
}

type Put struct {
	Cell
	value Slice
}

type Get struct {
	Cell
	value Slice
}

type Incr struct {
	Cell
	incrBy Slice
}

type Decr struct {
	Cell
	decrBy Slice
}

type Decay struct {
	Cell
	formula func(prevVal Slice)
}

type tableop interface {

	// Data Operation
	Get(gets ...Get) []Get
	Put(puts ...Put) error //Todo :  should create column on the fly ??
	Incr(incrs ...Incr)
	Decr(decrs ...Decr)
	Decay(decays ...Decay)

	// Admin Operations
	Open()
	Close()
	Disable() // Temporary Disable Table for any operation
	Archive() // Delete Table
}
