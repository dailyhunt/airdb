package sst

import (
	"bytes"
)

var numKeysPerBlock = 5

type Builder struct {
	counter   int // Number of keys written for the current block.
	sstBuffer *bytes.Buffer
}

func (b *Builder) Close() {

}
func (b *Builder) Finish() []byte {
	return nil

}

func NewSstBuilder() *Builder {

}
