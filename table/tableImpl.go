package table

import (
	"github.com/dailyhunt/airdb/region"
)

type KvTable struct {
	region region.Region
}

func (kv *KvTable) Close() {
	panic("implement me")
}

func (kv *KvTable) Drop() {
	panic("implement me")
}

func (kv *KvTable) Archive() {
	panic("implement me")
}

func (kv *KvTable) Put(put *Put) error {
	kv.region.Put(put)
}

func (kv *KvTable) Get() {
	panic("implement me")
}

func (kv *KvTable) Merge() {
	panic("implement me")
}

func (kv *KvTable) Add() {
	panic("implement me")
}

func (kv *KvTable) Decay() {
	panic("implement me")
}
