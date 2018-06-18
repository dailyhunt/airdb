package db

// TODO: manifest file format

type tableManifest struct {
	uuid  string
	name  string
	state PersistentState // table in addition could be archived, dropped too
}

// version => state => table metadata => num tables
type Manifest struct {
	version   string
	state     PersistentState
	numTables int
	tables    []tableManifest
}

func Read() (manifest *Manifest, err error) {

	return
}

func Write() {

}

func Compact() {

}

func AddTable() {

}

func UpdateTable() {

}