package table

import (
	"github.com/dailyhunt/airdb/operation"
	"github.com/dailyhunt/airdb/region"
)

type KvTable struct {
	Region region.Region
}

func (t *KvTable) AddRegionPeer(nodeId int64, url []byte) error {
	t.Region.AddPeer(nodeId, url)
	return nil
}

func (t *KvTable) Close() {
	panic("implement me")
}

func (t *KvTable) Drop() {
	panic("implement me")
}

func (t *KvTable) Archive() {
	panic("implement me")
}

func (t *KvTable) Put(put *operation.Put) error {
	// Todo: Validate key
	// Todo : Validate Value
	t.Region.Put(put)
	return nil
}

func (t *KvTable) Get() {
	panic("implement me")
}

func (t *KvTable) Merge() {
	panic("implement me")
}

func (t *KvTable) Add() {
	panic("implement me")
}

func (t *KvTable) Decay() {
	panic("implement me")
}
