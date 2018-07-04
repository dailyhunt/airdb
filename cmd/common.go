package cmd

import (
	"fmt"
	"github.com/dailyhunt/airdb/server"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"net"

	pb "github.com/dailyhunt/airdb/proto"
)

const (
//port = ":50051"
)

func StartAirdbServer() {

	// grpc config
	var grpcConfig grpcConfig
	err := viper.UnmarshalKey("grpc", &grpcConfig)
	if err != nil {
		panic(fmt.Errorf("Fatal error in reading 'grpc' config: %s \n", err))
	}

	// airdb options
	var dbOptions server.Options
	err = viper.UnmarshalKey("dbOptions", &dbOptions)
	if err != nil {
		panic(fmt.Errorf("Fatal error in reading 'dbOptions' config: %s \n", err))
	}

	port := grpcConfig.getPort()
	listener, err := net.Listen("tcp", port)
	if err != nil {
		panic(fmt.Errorf("could not create listener for grpc server %s", err))
	}

	grpcServer := grpc.NewServer()
	airdbSever := server.NewAirDBServer(dbOptions)
	pb.RegisterAirdbServer(grpcServer, airdbSever)

	if err = grpcServer.Serve(listener); err != nil {
		panic(fmt.Errorf("error while serving grpc server %s", err))
	}

}
