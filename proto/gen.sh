#!/usr/bin/env bash

protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf/protobuf --gogofaster_out=. operations.proto
protoc --gogofaster_out=plugins=grpc:. internal.proto