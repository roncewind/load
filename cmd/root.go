/*
Copyright © 2022 roncewind <dad@lynntribe.net>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/roncewind/load/input"
	"github.com/senzing/go-logging/logger"
	"github.com/senzing/go-logging/messagelogger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	delayInSeconds         string = "delay-in-seconds"
	envVarPrefix           string = "SENZING_TOOLS"
	envVarReplacerCharNew  string = "_"
	envVarReplacerCharOld  string = "-"
	inputFileTypeParameter string = "input-file-type"
	inputURLParameter      string = "input-url"
	logLevelParameter      string = "log-level"
	withInfoParameter      string = "with-info"
)

var (
	buildIteration string = "0"
	buildVersion   string = "0.0.0"
	programName    string = "load"
)

var (
	cfgFile    string
	delay      int    = 0
	exchange   string = "senzing"
	fileType   string //TODO: load from file
	inputQueue string = "senzing_input"
	inputURL   string // read from this URL, could be a file or a queue
	logLevel   string = "info"
	msglog     messagelogger.MessageLoggerInterface
	withInfo   bool = false
)

// load is 6201:  https://github.com/Senzing/knowledge-base/blob/main/lists/senzing-product-ids.md
const productIdentifier = 6201

var idMessages = map[int]string{
	1:    "Viper key list:",
	2:    "  - %s = %s",
	11:   "%s has a score of %d.",
	999:  "A test of INFO.",
	1000: "A test of WARN.",
	2000: "A test of ERROR.",
	2001: "Config file found, but not loaded",
}

// ----------------------------------------------------------------------------
// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "load",
	Short: "Load records into Senzing",
	Long: `TODO: Load records from somewhere...

	For example:

load --input-url "amqp://guest:guest@192.168.6.96:5672?exchange=senzing-rabbitmq-exchange&queue-name=senzing-rabbitmq-queue&routing-key=senzing.records"
`,

	Run: func(cmd *cobra.Command, args []string) {
		if msglog.IsInfo() {
			msglog.Log(1, logger.LevelInfo)
			for _, key := range viper.AllKeys() {
				// msglog.Log(2, key, viper.Get(key), logger.LevelInfo)
				fmt.Println(key, ":", viper.Get(key))
			}
		}
		fmt.Println(time.Now(), "Sleep for 60 seconds to let RabbitMQ and Postgres settle down and come up.")
		time.Sleep(time.Duration(delay) * time.Second)
		if !input.Read(inputURL, logLevel) {
			cmd.Help()
		}

	},
}

// ----------------------------------------------------------------------------
// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// ----------------------------------------------------------------------------
func init() {
	msglog, _ = messagelogger.NewSenzingLogger(productIdentifier, idMessages)
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.senzing/config.yaml)")

	// local flags for load
	RootCmd.Flags().IntVarP(&delay, delayInSeconds, "", 0, "time to wait before start of processing")
	viper.BindPFlag(delayInSeconds, RootCmd.Flags().Lookup(delayInSeconds))

	RootCmd.Flags().StringVarP(&fileType, inputFileTypeParameter, "", "", "file type override")
	viper.BindPFlag(inputFileTypeParameter, RootCmd.Flags().Lookup(inputFileTypeParameter))

	RootCmd.Flags().StringVarP(&inputURL, inputURLParameter, "i", "", "input location")
	viper.BindPFlag(inputURLParameter, RootCmd.Flags().Lookup(inputURLParameter))

	RootCmd.Flags().StringVarP(&logLevel, logLevelParameter, "", "", "set the logging level, default Error")
	viper.BindPFlag(logLevelParameter, RootCmd.Flags().Lookup(logLevelParameter))

	RootCmd.Flags().BoolP(withInfoParameter, "", false, "set to add record withInfo")
	viper.BindPFlag(withInfoParameter, RootCmd.Flags().Lookup(withInfoParameter))
}

// ----------------------------------------------------------------------------
// initConfig reads in config file and ENV variables if set.
// Config precedence:
// - cmdline args
// - env vars
// - config file
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in <home directory>/.senzing with name "config" (without extension).
		viper.AddConfigPath(home + "/.senzing-tools")
		viper.AddConfigPath(home)
		viper.AddConfigPath("/etc/senzing-tools")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error
		} else {
			// Config file was found but another error was produced
			msglog.Log(2001, err)
		}
	}
	viper.AutomaticEnv() // read in environment variables that match
	// all env vars should be prefixed with "SENZING_TOOLS_"
	replacer := strings.NewReplacer(envVarReplacerCharOld, envVarReplacerCharNew)
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix(envVarPrefix)
	viper.BindEnv(delayInSeconds)
	viper.BindEnv(inputFileTypeParameter)
	viper.BindEnv(inputURLParameter)
	viper.BindEnv(logLevelParameter)
	viper.BindEnv(withInfoParameter)

	// cmdline args should get set in viper, but for some reason that's
	// not happening when called from senzing-tools, this is the work around:
	if delay > 0 {
		viper.Set(delayInSeconds, delay)
	}
	if len(fileType) > 0 {
		viper.Set(inputFileTypeParameter, fileType)
	}
	if len(inputURL) > 0 {
		viper.Set(inputURLParameter, inputURL)
	}
	if len(logLevel) > 0 {
		viper.Set(logLevelParameter, logLevel)
	}
	if withInfo {
		viper.Set(withInfoParameter, withInfo)
	}

	viper.SetDefault(delayInSeconds, 0)
	viper.SetDefault(logLevelParameter, "error")
	viper.SetDefault(withInfoParameter, false)

	// setup local variables, in case they came from a config file
	//TODO:  why do I have to do this?  env vars and cmdline params get mapped
	//  automatically, this is only IF the var is in the config file.
	//  am i missing a way to bind config file vars to local vars?
	delay = viper.GetInt(delayInSeconds)
	fileType = viper.GetString(inputFileTypeParameter)
	inputURL = viper.GetString(inputURLParameter)
	logLevel = viper.GetString(logLevelParameter)
	withInfo = viper.GetBool(withInfoParameter)

	msglog.SetLogLevelFromString(logLevel)
}
