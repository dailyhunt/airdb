package server

import (
	"time"
)

type KvCommand struct {
	Op    string
	Key   string
	Value string
	Epoch time.Time
}
