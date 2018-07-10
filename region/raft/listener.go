package raft

import (
	"errors"
	"net"
	"time"
)

type raftListener struct {
	*net.TCPListener
	stopStream <-chan struct{}
}

func newRaftListener(addr string, stopc <-chan struct{}) (*raftListener, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &raftListener{ln.(*net.TCPListener), stopc}, nil
}

func (ln raftListener) Accept() (c net.Conn, err error) {
	connectionStream := make(chan *net.TCPConn, 1)
	errorStream := make(chan error, 1)
	go func() {
		tc, err := ln.AcceptTCP()
		if err != nil {
			errorStream <- err
			return
		}
		connectionStream <- tc
	}()
	select {
	case <-ln.stopStream:
		return nil, errors.New("server stopped")
	case err := <-errorStream:
		return nil, err
	case tc := <-connectionStream:
		tc.SetKeepAlive(true)
		tc.SetKeepAlivePeriod(3 * time.Minute)
		return tc, nil
	}
}
