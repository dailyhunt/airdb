package table

import (
	"github.com/gin-gonic/gin/json"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenFunction(t *testing.T) {
	//logger := logrus.New()
	log.SetLevel(log.ErrorLevel)
	baseDir, _ := os.Getwd()
	dbDir := "tmpdb"
	tableName := "t1"
	createManifestFile(baseDir, dbDir, tableName, t)

	path := filepath.Join(baseDir, dbDir, tableName)
	table, err := OpenTable(path)
	if err != nil {
		t.Errorf("could not open table at %s ", path)
	}

	if table.Name() != tableName {
		t.Errorf("table name expected %s but got %s", table.Name(), tableName)
	}

	os.RemoveAll(dbDir)

}

func createManifestFile(baseDir string, dbDir string, tableName string, t *testing.T) {
	testDir := filepath.Join(baseDir, dbDir, tableName)
	manifestFile := filepath.Join(testDir, "manifest.json")
	os.RemoveAll(testDir)
	err := os.MkdirAll(testDir, os.ModePerm)
	if err != nil {
		t.Error("Error while create baseDir ", err)
	}
	mf := Manifest{
		Version:       1,
		ActiveRegions: []int64{},
		Options: Options{
			Name:           tableName,
			Path:           testDir,
			SstDir:         filepath.Join(testDir, "sst"),
			WalDir:         testDir,
			MaxRegions:     1,
			RegionSeqId:    0,
			CurrentRegions: 0,
		},
	}
	bytes, _ := json.Marshal(mf)
	err = ioutil.WriteFile(manifestFile, bytes, 777)
	if err != nil {
		t.Error("Error while writing manifest file")
	}

}
