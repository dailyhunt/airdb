package table

import (
	"context"
	"github.com/dailyhunt/airdb/mutation"
	r "github.com/dailyhunt/airdb/region"
	log "github.com/sirupsen/logrus"
)

type tableImpl struct {
	name     string
	regions  map[int]r.Region
	manifest *Manifest
	opts     Options
}

func NewTableImpl(m *Manifest) *tableImpl {
	t := &tableImpl{
		name:     m.Options.Name,
		manifest: m,
		opts:     m.Options,
	}
	return t
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

func (t *tableImpl) Mutate(ctx context.Context, data []byte) error {
	m, _ := mutation.Decode(data)
	log.Infof("proposing mutation to region ... %s ", string(m.Key))
	region := t.GetRegion(m.Key)
	region.Mutate(ctx, data)
	return nil
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

func (t *tableImpl) GetRegion(key []byte) r.Region {
	region, ok := t.regions[999]
	if !ok {
		log.Fatalf("no region found with id %v ", 999, " available regions are (%v)", t.regions)
	}
	return region
}

func (t *tableImpl) FlushManifest() {

}
