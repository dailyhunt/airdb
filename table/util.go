package table

import (
	"encoding/json"
	"fmt"
	"github.com/dailyhunt/airdb/utils/fileUtils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"path/filepath"
)

func readManifestFile(tableDir string) (*Manifest, error) {
	file := filepath.Join(tableDir, "manifest.json")
	exists, err := fileUtils.Exists(file)
	if !exists || err != nil {
		return nil, errors.New(fmt.Sprintf("manifest file does not exists at %s ", tableDir))
	}

	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var manifest Manifest

	err = json.Unmarshal(bytes, &manifest)
	if err != nil {
		return nil, err
	}

	log.Infof("successfully read manifest for table %s : manifest %s  ", tableDir, string(bytes))

	return &manifest, nil

}
