package server

import (
	"context"
	"fmt"
	"github.com/dailyhunt/airdb/db"
	"github.com/dailyhunt/airdb/proto"
	"github.com/dailyhunt/airdb/table"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	opts Options
	db   db.DB
}

type Options struct {
	dbPath string
}

func NewServer(opts Options) *Server {
	d, err := db.Open(opts.dbPath)
	if err != nil {
		// Todo : Add Logging
	}
	return &Server{
		opts: opts,
		db:   d,
	}
}

func (s *Server) getTable(name string) (table.Table, error) {
	// Todo : Check if table exists
	t, err := s.db.GetTable(name)
	if err != nil {
		log.Error(fmt.Sprintf("Error while getting table : %s", name))
		return nil, err
	}
	return t, nil
}

func (s *Server) Put(ctx context.Context, req *server.OpRequest) (*server.OpResponse, error) {
	if err := validatePutRequest(req); err != nil {
		return nil, err
	}

	t, err := s.getTable(req.GetTable())
	if err != nil {
		return nil, err
	}

	err = t.Put(ctx, req.GetMutation())
	if err != nil {
		log.Error(fmt.Sprintf("Error while executing PUT on table : %s", req.GetTable()))
		return nil, err
	}

	// Todo : Response
	return &server.OpResponse{}, nil

}

func (s *Server) Get(ctx context.Context, req *server.OpRequest) (*server.OpResponse, error) {
	panic("implement me")
}
