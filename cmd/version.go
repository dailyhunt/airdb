package cmd

import (
	"github.com/spf13/cobra"
	"fmt"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of AirDB",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: get version from config file
		fmt.Println("0.0.1")
	},
}
