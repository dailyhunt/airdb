package cmd

import (
	"context"
	"errors"
	"time"

	"fmt"

	"github.com/dailyhunt/airdb/proto"
	pb "github.com/dailyhunt/airdb/proto"
	logger "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var putCmd = &cobra.Command{
	Use:   "put",
	Short: "put <table> <key> <family:column> <value>",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 4 {
			return errors.New("invalid number of arguments passed . Required argument length= 4")
		}
		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("args ", args)

		table := args[0]

		put := &pb.Put{
			Key:   []byte(args[1]),
			Col:   []byte(args[2]),
			Value: []byte(args[3]),
			Epoch: uint64(time.Now().Unix()),
		}

		dAtA, err := put.Marshal()

		if err != nil {
			logger.Fatal("error while marshalling put")
		}

		req := &proto.OpRequest{Table: table, ReqBody: dAtA}

		logger.WithFields(logger.Fields{"table": table, "key": args[1], "column": args[2]}).Info("put operation")

		client, err := CreateAirdbClientConn()
		if err != nil {
			logger.Fatal(err)
		}

		res, err := client.Put(context.Background(), req)

		if err != nil {
			logger.Fatal(err)
		}

		logger.WithFields(logger.Fields{"key": args[1], "response": string(res.Value)}).Info("put success .")

	},
}

func init() {
	rootCmd.AddCommand(putCmd)

}
