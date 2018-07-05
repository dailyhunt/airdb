package cmd

import (
	"context"
	"errors"
	"github.com/dailyhunt/airdb/mutation"
	"github.com/dailyhunt/airdb/proto"
	logger "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
	"time"
)

var putCmd = &cobra.Command{
	Use:   "put <table> <key> <family:column> <value>",
	Short: "Put a cell and value",
	Long:  `Put a <key> <family:column> <value> pair to <table> on airdb.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 4 {
			return errors.New("invalid number of arguments passed . Required argument length= 4")
		}
		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {

		cell := strings.Split(args[2], ":")

		m := &mutation.Mutation{
			Key:          []byte(args[1]),
			Family:       []byte(cell[0]),
			Col:          []byte(cell[1]),
			Value:        []byte(args[3]),
			Timestamp:    uint64(time.Now().Unix()),
			MutationType: mutation.PUT,
		}

		req := &proto.OpRequest{
			Table:   args[0],
			ReqBody: mutation.Encode(m),
		}

		logger.WithFields(logger.Fields{"table": args[0], "key": args[1], "family": cell[0], "column": cell[1]}).Info("put operation")

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
