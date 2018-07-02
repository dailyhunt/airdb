package server

import (
	"context"
	"github.com/dailyhunt/airdb/db"
	"github.com/dailyhunt/airdb/proto"
	"github.com/dailyhunt/airdb/table"
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

func (s *Server) getTable(string) table.Table {

}

func (s *Server) Put(ctx context.Context, req *server.OpRequest) (*server.OpResponse, error) {
	if err := validatePutRequest(req); err != nil {
		return nil, err
	}

	// Todo : Check if table exists

	table := s.getTable(req.GetTable())
	table.Put(ctx, req.GetMutation())

}

func (s *Server) Get(ctx context.Context, req *server.OpRequest) (*server.OpResponse, error) {
	panic("implement me")
}
