package cmd

import (
	"fmt"

	logger "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"strings"
)

const AppName = "airdb"

var cfgFile string
var raftPort int
var raftId int64

var rootCmd = &cobra.Command{
	Use:   "airdb",
	Short: "AirDB is very fast key value LSM tree database",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
		logger.Info("Starting airdb server ..... ")
		viper.Set("raft.port", raftPort)
		viper.Set("raft.id", raftId)
		StartAirdbServer()

	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// TODO: mention default search order
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file")
	rootCmd.PersistentFlags().IntVarP(&raftPort, "raftPort", "r", 12000, "raft port")
	rootCmd.PersistentFlags().Int64VarP(&raftId, "raftId", "n", 12000, "raft Id")

	//
	// enable commandline flags
	//
	//pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	//
	// enable environment
	//
	viper.SetEnvPrefix(AppName)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}

func initConfig() {
	// Don't forget to read config either from cfgFile or from home directory!
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		//
		// Config file
		//
		env := os.Getenv("env")
		if env == "" {
			env = "local"
		}

		viper.SetConfigName(fmt.Sprintf("config-%s", env))
		viper.SetConfigType("yaml")
		viper.AddConfigPath("./config")
		viper.AddConfigPath("./airdb/config")
		viper.AddConfigPath("$HOME/" + AppName + "/config")
		viper.AddConfigPath("/usr/local/etc/" + AppName + "/config")
	}

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Fatal error in reading config file: %s \n", err))
	}

	configureLogger()
}

func configureLogger() {

	var logConfig logConfig

	err := viper.UnmarshalKey("log", &logConfig)
	if err != nil {
		panic(fmt.Errorf("Fatal error in reading 'log' config: %s \n", err))
	}

	if logConfig.Format == JSON {
		logger.SetFormatter(&logger.JSONFormatter{})
	} else if logConfig.Format == TEXT {
		logger.SetFormatter(&logger.TextFormatter{})
	} else {
		panic(fmt.Errorf("Unsupported log format: %s ! Supported Values: json, text \n", logConfig.Format))
	}

	if logConfig.Output == STDOUT {

		logger.SetOutput(os.Stdout)

	} else if logConfig.Output == FILE {
		logger.SetOutput(&lumberjack.Logger{
			Filename:   logConfig.File.Filename,
			MaxSize:    logConfig.File.MaxSize, // megabytes
			MaxBackups: logConfig.File.MaxBackups,
			MaxAge:     logConfig.File.MaxAge,   //days
			Compress:   logConfig.File.Compress, // disabled by default
		})
	} else {
		panic(fmt.Errorf("Unsupported log output: %s ! Supported Values: stdout, file \n", logConfig.Output))
	}

	logLevel, err := logger.ParseLevel(logConfig.Level)
	if err != nil {
		panic(fmt.Errorf("Unsupported log level: %s ! Supported Values: panic, fatal, error, warn, warning, debug, info \n", logConfig.Level))
	}

	logger.SetLevel(logLevel)
}
