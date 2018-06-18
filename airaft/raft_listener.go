package airaft

import (
	"errors"
	"net"
	"time"
)

// stoppableListener sets TCP keep-alive timeouts on accepted
// connections and waits on stopChan message
type raftListener struct {
	*net.TCPListener
	stopChan <-chan struct{}
}

func newRaftListener(addr string, stopChan <-chan struct{}) (*raftListener, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &raftListener{ln.(*net.TCPListener), stopChan}, nil
}
func (ln *raftListener) Accept() (c net.Conn, err error) {
	connChan := make(chan *net.TCPConn, 1)
	errorChan := make(chan error, 1)

	go func() {

		// AcceptTCP accepts the next incoming call and returns the new connection.
		tc, err := ln.AcceptTCP()
		if err != nil {
			errorChan <- err
			return
		}
		connChan <- tc

	}()

	select {
	case <-ln.stopChan:
		return nil, errors.New("server stopped")
	case err := <-errorChan:
		return nil, err
	case tc := <-connChan:
		// Todo (sohan) - take from config
		tc.SetKeepAlive(true)
		tc.SetKeepAlivePeriod(3 * time.Minute)
		return tc, nil

	}
}
