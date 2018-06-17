package table

import (
	"time"
)

type Put struct {
	K string
	V string
	T time.Time
}

type Get struct {
	K string
	V string
}
