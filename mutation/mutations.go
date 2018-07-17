package mutation

import (
	"encoding/json"
)

type Type uint16

const (
	PUT Type = 1
	GET Type = 2
)

var typeNames = map[Type]string{
	1: "PUT",
	2: "GET",
}
var typeValues = map[string]Type{
	"PUT": 1,
	"GET": 2,
}

func (t Type) String() string {
	// Todo : Check for casting
	return typeNames[t]
}

func (t Type) Type(name string) Type {
	return typeValues[name]
}

type DataType uint16

const (
	SINGLE DataType = 1
	LIST   DataType = 2
)

var dataTypeNames = map[DataType]string{
	1: "SINGLE",
	2: "LIST",
}
var dataTypeValues = map[string]DataType{
	"SINGLE": 1,
	"LIST":   2,
}

func (t DataType) String() string {
	// Todo : Check for casting
	return dataTypeNames[t]
}

func (t DataType) Type(name string) DataType {
	return dataTypeValues[name]
}

// Mutation
type Mutation struct {
	Key          []byte
	Family       []byte
	Col          []byte
	Value        []byte
	Timestamp    uint64
	MutationType Type
}

func (m *Mutation) String() string {
	mp := make(map[string]interface{})
	mp["key"] = string(m.Key)
	mp["fam"] = string(m.Family)
	mp["col"] = string(m.Col)
	mp["val"] = string(m.Value)
	mp["tim"] = m.Timestamp
	mp["typ"] = m.MutationType.String()
	bytes, _ := json.Marshal(mp)
	return string(bytes)
}
