package db

import (
	"github.com/dailyhunt/airdb/table"
	"github.com/dailyhunt/airdb/utils/fileUtils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
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

type Handle struct {
	lock            *flock.Flock
	path            string
	persistentState PersistentState
	runState        RunState
	tables          map[string]table.Table
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

	// open all active tables

	return
}

// check if db exists
// if exists then panic
// if not exists then create it and take a lock
// if yes, then take the lock. else panic
// internally, db open would require loading all tables metadata and opening tables
func (db *Handle) Create(config Options) (err error) {
	db.path = config.Path

	exists, err := fileUtils.Exists(config.Path)
	if err != nil {
		err = errors.Wrap(err, "Error in creating DB at path ("+config.Path+")")
		return
	}

	if exists {
		if err = fileUtils.AssertDir(config.Path); err != nil {
			err = errors.Wrap(err, "Error in creating DB at path ("+config.Path+")")
			return
		}

		if err = fileUtils.AssertEmpty(config.Path); err != nil {
			err = errors.Wrap(err, "Error in creating DB at path ("+config.Path+")")
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
	t, ok := db.tables[name]
	if !ok {
		e := errors.Errorf("table not found :  %s", name)
		return nil, e
	}
	return t, nil
}

func (db *Handle) CreateTable(config table.Options) (table table.Table, err error) {
	return
}

func (db *Handle) DropTable(name string) (err error) {
	return
}

func (db *Handle) ArchiveTable(name string) (err error) {
	return
}

func (db *Handle) AddTestTable() {
	db.tables = make(map[string]table.Table)
	path := "/home/sohanvir/softwares/go/src/github.com/dailyhunt/adb/t1"
	t, err := table.OpenTable(path)
	if err != nil {
		panic(err)
	}

	db.tables[t.Name()] = t
	log.Infof("successfully opened table %s ", t.Name())
}
