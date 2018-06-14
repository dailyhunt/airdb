package server

import (
	"time"
)

const (
	OP_SET    = "SET"
	OP_GET    = "GET"
	OP_DELETE = "DELETE"
)

type KvCommand struct {
	Op    string
	Key   string
	Value string
	Epoch time.Time
}
