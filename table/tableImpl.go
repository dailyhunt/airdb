package table

import (
	"context"
	"github.com/dailyhunt/airdb/proto"
)

type tableImpl struct {
}

func (tableImpl) Close() {
	panic("implement me")
}

func (tableImpl) Drop() {
	panic("implement me")
}

func (tableImpl) Archive() {
	panic("implement me")
}

func (tableImpl) Put(ctx context.Context, mutation *server.Mutation) {
	panic("implement me")
}

func (tableImpl) Get() {
	panic("implement me")
}

func (tableImpl) Merge() {
	panic("implement me")
}

func (tableImpl) Add() {
	panic("implement me")
}

func (tableImpl) Decay() {
	panic("implement me")
}
