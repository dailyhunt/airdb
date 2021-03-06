package db

import (
	"github.com/dailyhunt/airdb/table"
	"github.com/dailyhunt/airdb/utils/fileUtils"
	"github.com/pkg/errors"
	"github.com/theckman/go-flock"
	"path/filepath"
)

//
// Directory layout
// (db root)
//		- db manifest file (contains active tables etc)
//		- wal
//		- (table root, multiple such tables)
//				- table manifest file (contains active region etc)
//				- (region root, multiple such regions)
//						- region manifest file
//						- sst files
//						- vlog files
//

const LockFile = "airdb.lock"
const ManifestFile = "airdb.manifest"

type PersistentState int

const (
	CreatedPState     PersistentState = iota
	InitialisedPState
)

type RunState int

const (
	CreatingRState     RunState = iota
	CreatedRState
	InitialisingRState
	InitialisedRState
	StartingRState
	StartedRState
)

type Handle struct {
	lock            *flock.Flock
	path            string
	persistentState PersistentState
	runState        RunState
	tables          []table.Table // list of tables
}

// check if db exists
// if not, then panic
// if exists, then see if lock can be taken
// if yes, then take the lock. else panic
// read manifest and populate handle
func (db *Handle) Open(path string) (err error) {
	db.path = path

	err = fileUtils.AssertExists(path)
	if err != nil {
		err = errors.Wrap(err, "Error in opening DB at path ("+path+")")
		return
	}

	err = fileUtils.AssertDir(path)
	if err != nil {
		err = errors.Wrap(err, "Error in opening DB at path ("+path+")")
		return
	}

	// TODO: check if manifest file exists or not

	// take a lock
	db.Lock()

	// open and load manifest file

	return
}

// check if db exists
// if exists then panic
// if not exists then create it and take a lock
// if yes, then take the lock. else panic
// internally, db open would require loading all tables metadata and opening tables
func (db *Handle) Create(config CreateConfig) (err error) {
	db.path = config.path

	exists, err := fileUtils.Exists(config.path)
	if err != nil {
		err = errors.Wrap(err, "Error in creating DB at path ("+config.path+")")
		return
	}

	if exists {
		if err = fileUtils.AssertDir(config.path); err != nil {
			err = errors.Wrap(err, "Error in creating DB at path ("+config.path+")")
			return
		}

		if err = fileUtils.AssertEmpty(config.path); err != nil {
			err = errors.Wrap(err, "Error in creating DB at path ("+config.path+")")
			return
		}
	} else {
		// create directory
	}

	// take a lock
	db.Lock()

	// create manifest file

	return
}

//func (db *Handle) Init(config InitConfig) (err error) {
//	return
//}

func (db *Handle) Close() (err error) {
	// close all underlying tables

	// remove lock
	db.Unlock()

	return
}

func (db *Handle) Lock() (err error) {
	lock := flock.NewFlock(filepath.Join(db.path, LockFile))

	locked, err := lock.TryLock()

	if err != nil {
		// handle locking error
	}

	if !locked {

	}

	db.lock = lock

	return
}

func (db *Handle) Unlock() (err error) {
	if db.lock != nil {
		// do work
		db.lock.Unlock()
	}

	return
}

func (db *Handle) ListTables(inactive bool) (tables []table.Table, err error) {
	return
}

func (db *Handle) GetTable(name string) (table table.Table, err error) {
	return
}

func (db *Handle) CreateTable(config table.CreateConfig) (table table.Table, err error) {
	return
}

func (db *Handle) DropTable(name string) (err error) {
	return
}

func (db *Handle) ArchiveTable(name string) (err error) {
	return
}
