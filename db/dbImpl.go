package db

import "github.com/dailyhunt/airdb/table"

//
// Directory layout
// (DB root)
//		- DB manifest file (contains active tables etc)
//		- wal
//		- (table root, multiple such tables)
//				- table manifest file (contains active region etc)
//				- (region root, multiple such regions)
//						- region manifest file
//						- sst files
//						- vlog files
//

type PersistentState int

const (
	CreatedPState PersistentState = iota
	InitialisedPState
)

type RunState int

const (
	CreatingRState RunState = iota
	CreatedRState
	InitialisingRState
	InitialisedRState
	StartingRState
	StartedRState
)

// TODO: manifest file format

type Handle struct {
	// TODO: directory lock
	persistentState PersistentState
	runState        RunState
	tables          map[string]*table.Table // map of tables
}

// find DB manifest file at the path.
func (db *Handle) Open(path string) {
	// check if DB exists
	// if not, then panic
	// if exists, then see if lock can be taken
	// if yes, then take the lock. else panic
	// read manifest and populate handle
}

func (db *Handle) Create() {
	// check if DB exists
	// if exists then panic
	// if not exists then create it and take a lock
	// if yes, then take the lock. else panic
	// internally, DB open would require loading all tables metadata and opening tables
}

func (db *Handle) Init() {

}

func (db *Handle) Close() {

}

func (db *Handle) ListTables() {

}

func (db *Handle) AddTable() {

}

func (db *Handle) DropTable() {

}

func (db *Handle) ArchiveTable() {

}

// version => state => num tables => table metadata
type manifest struct {
	version string
	state   PersistentState
}

func readManifest() {

}

func persistManifest() {

}
