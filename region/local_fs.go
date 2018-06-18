package region

import (
	"github.com/dailyhunt/airdb/operation"
)

type LocalFs struct {
	*Replica
}

func (fs *LocalFs) Close() {
	panic("implement me")
}

func (fs *LocalFs) Drop() {
	panic("implement me")
}

func (fs *LocalFs) Archive() {
	panic("implement me")
}

func (fs *LocalFs) Put(put *operation.Put) error {
	panic("implement me")
}

func (fs *LocalFs) Get() {
	panic("implement me")
}

func (fs *LocalFs) Merge() {
	panic("implement me")
}

func (fs *LocalFs) Add() {
	panic("implement me")
}

func (fs *LocalFs) Decay() {
	panic("implement me")
}
