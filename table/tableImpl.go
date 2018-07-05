package table

import (
	"context"
)

type tableImpl struct {
}

func (t *tableImpl) Open(option *Options) {
	panic("implement me")
}

func (t *tableImpl) Close() {
	panic("implement me")
}

func (t *tableImpl) Drop() {
	panic("implement me")
}

func (t *tableImpl) Archive() {
	panic("implement me")
}

func (t *tableImpl) Put(ctx context.Context, data []byte) error {
	panic("implement me")
}

func (t *tableImpl) Get() {
	panic("implement me")
}

func (t *tableImpl) Merge() {
	panic("implement me")
}

func (t *tableImpl) Add() {
	panic("implement me")
}

func (t *tableImpl) Decay() {
	panic("implement me")
}
