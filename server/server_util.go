package server

import (
	pb "github.com/dailyhunt/airdb/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func validatePutRequest(req *pb.OpRequest) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "Req cannot be empty.")
	}

	if req.GetTable() == "" {
		return status.Error(codes.InvalidArgument, "Table name can not be empty")
	}

	// Todo : Other validations

	return nil
}
