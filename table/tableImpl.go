package table

import (
	"context"
	r "github.com/dailyhunt/airdb/region"
)

type tableImpl struct {
	name     string
	regions  map[int64]r.Region
	manifest *Manifest
	opts     Options
}

func (t *tableImpl) Name() string {
	return t.name
}

func (t *tableImpl) Open(option *Options) {
	panic("implement me")
}

func (t *tableImpl) Close() {
	t.FlushManifest()
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
func (t *tableImpl) FlushManifest() {

}
