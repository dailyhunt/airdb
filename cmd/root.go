package cmd

import (
	"fmt"
	"github.com/dailyhunt/airdb/db"
	"github.com/dailyhunt/airdb/server"
	"github.com/onrik/logrus/filename"
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
var nodeId int
var peers string

var rootCmd = &cobra.Command{
	Use:   "airdb",
	Short: "AirDB is very fast key value LSM tree database",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// start HTTP Server
		// Todo: Create store first and pass to all api servers
		fmt.Println("Id ", nodeId)
		fmt.Println("Cluster ", peers)
		db, err := db.OpenForDebug(nodeId, peers)

		// Todo: (sohan) add proper logging
		if err != nil {
			logger.Error("Error while opening database at dir ", "dummy dir")
		}
		server.StartHTTPServer(db)
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
	rootCmd.PersistentFlags().StringVarP(&peers, "peers", "p", "", "peers")
	rootCmd.PersistentFlags().IntVarP(&nodeId, "nodeId", "n", 0, "node id")

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
		fmt.Print("Config is not empty")
	} else {
		//
		// Config file
		//
		env := os.Getenv("env")
		if env == "" {
			env = "local"
			viper.AddConfigPath("./airdb/config")
		}

		viper.SetConfigName(fmt.Sprintf("config-%s", env))
		viper.SetConfigType("yaml")
		viper.AddConfigPath("./config")
		viper.AddConfigPath("$HOME/" + AppName + "/config")
		viper.AddConfigPath("/usr/local/etc/" + AppName + "/config")
	}

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Fatal error in reading config file: %s \n", err))
	}

	configureLogger()
}

func configureLogger() {
	type logFormat string

	const (
		JSON logFormat = "json"
		TEXT           = "text"
	)

	type logOutput string

	const (
		FILE   logOutput = "file"
		STDOUT           = "stdout"
	)

	var logConfig struct {
		Level  string
		Format logFormat
		Output logOutput
		File   struct {
			Filename   string
			MaxSize    int
			Compress   bool
			MaxBackups int
			MaxAge     int
		}
	}

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

	//logLevel, err := logger.ParseLevel(logConfig.Level)
	logLevel, err := logger.ParseLevel("debug")
	if err != nil {
		panic(fmt.Errorf("Unsupported log level: %s ! Supported Values: panic, fatal, error, warn, warning, debug, info \n", logConfig.Level))
	}

	logger.SetLevel(logLevel)
	logger.AddHook(filename.NewHook())
}
