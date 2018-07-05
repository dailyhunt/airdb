package cmd

import (
	"fmt"
	"github.com/dailyhunt/airdb/server"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"net"

	pb "github.com/dailyhunt/airdb/proto"
	logger "github.com/sirupsen/logrus"
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

	logger.WithFields(logger.Fields{"port": grpcConfig.Port, "dbPath": dbOptions.DbPath}).Info("airdb arguments and options.")

	grpcServer := grpc.NewServer()
	airdbSever := server.NewAirDBServer(dbOptions)
	pb.RegisterAirdbServer(grpcServer, airdbSever)

	if err = grpcServer.Serve(listener); err != nil {
		panic(fmt.Errorf("error while serving grpc server %s", err))
	}

}

func CreateAirdbClientConn() (pb.AirdbClient, error) {
	port := viper.GetString("grpc.port")
	if port == "" {
		logger.Fatalf("Empty grpc port provided %v", viper.AllKeys())
	}

	conn, err := grpc.Dial(fmt.Sprintf("localhost:%s", port), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	client := pb.NewAirdbClient(conn)
	return client, nil

}
