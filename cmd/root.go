/*
Copyright © 2022 roncewind <dad@lynntribe.net>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/senzing/go-logging/messagelogger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	defaultDelayInSeconds int    = 0
	defaultFileType       string = ""
	defaultInputURL       string = ""
	defaultOutputURL      string = ""
	defaultLogLevel       string = "error"
	Use                   string = "load"
	Short                 string = "Load records into Senzing."
	Long                  string = `
	Welcome to load!
	This tool will load records into Senzing. It validates the records conform to the Generic Entity Specification.

	For example:

	load --input-url "amqp://guest:guest@192.168.6.96:5672"
	load --input-url "https://public-read-access.s3.amazonaws.com/TestDataSets/SenzingTruthSet/truth-set-3.0.0.jsonl"
`
)
const (
	delayInSeconds         string = "delay-in-seconds"
	engineConfigParameter  string = "engine-configuration-json"
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
	cfgFile string
	delay   int = 0
	// engineConfigJson string
	exchange   string = "senzing"
	fileType   string //TODO: load from file
	inputQueue string = "senzing_input"
	// inputURL         string // read from this URL, could be a file or a queue
	// logLevel         string = "info"
	msglog   messagelogger.MessageLoggerInterface
	withInfo bool = false
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
	Use:   Use,
	Short: Short,
	Long:  Long,
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag(delayInSeconds, cmd.Flags().Lookup(delayInSeconds))
		viper.BindPFlag(engineConfigParameter, cmd.Flags().Lookup(engineConfigParameter))
		viper.BindPFlag(inputFileTypeParameter, cmd.Flags().Lookup(inputFileTypeParameter))
		viper.BindPFlag(inputURLParameter, cmd.Flags().Lookup(inputURLParameter))
		viper.BindPFlag(logLevelParameter, cmd.Flags().Lookup(logLevelParameter))
		viper.BindPFlag(withInfoParameter, cmd.Flags().Lookup(withInfoParameter))
	},
	Run: func(cmd *cobra.Command, args []string) {
		// if msglog.IsInfo() {
		// msglog.Log(1, logger.LevelInfo)
		fmt.Println("Viper keys")
		for _, key := range viper.AllKeys() {
			// msglog.Log(2, key, viper.Get(key), logger.LevelInfo)
			fmt.Println(key, ":", viper.Get(key))
		}
		// }
		fmt.Println(time.Now(), "Sleep for ", delay, " seconds to let queues and database settle down and come up.")
		time.Sleep(time.Duration(delay) * time.Second)

		// ctx := context.Background()

		// loader := &loader.LoaderImpl{
		// 	InputURL:         viper.GetString(inputURLParameter),
		// 	LogLevel:         viper.GetString(logLevelParameter),
		// 	EngineConfigJson: viper.GetString(engineConfigParameter),
		// }

		// if !loader.Load(ctx) {
		// 	cmd.Help()
		// }

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
	RootCmd.Flags().IntVar(&delay, delayInSeconds, 0, "time to wait before start of processing")
	RootCmd.Flags().StringVar(&engineConfigJson, engineConfigParameter, "", "Senzing engine configuration JSON")
	RootCmd.Flags().StringVar(&fileType, inputFileTypeParameter, "", "file type override")
	RootCmd.Flags().StringVarP(&inputURL, inputURLParameter, "i", "", "input location")
	RootCmd.Flags().StringVar(&logLevel, logLevelParameter, "", "set the logging level, default Error")
	RootCmd.Flags().Bool(withInfoParameter, false, "set to add record withInfo")
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
	viper.BindEnv(engineConfigParameter)
	viper.BindEnv(inputFileTypeParameter)
	viper.BindEnv(inputURLParameter)
	viper.BindEnv(logLevelParameter)
	viper.BindEnv(withInfoParameter)

	viper.SetDefault(delayInSeconds, 0)
	viper.SetDefault(logLevelParameter, "error")
	viper.SetDefault(withInfoParameter, false)

	// setup local variables, in case they came from a config file
	//TODO:  why do I have to do this?  env vars and cmdline params get mapped
	//  automatically, this is only IF the var is in the config file.
	//  am i missing a way to bind config file vars to local vars?
	delay = viper.GetInt(delayInSeconds)
	// engineConfigJson = viper.GetString(engineConfigParameter)
	fileType = viper.GetString(inputFileTypeParameter)
	// inputURL = viper.GetString(inputURLParameter)
	// logLevel = viper.GetString(logLevelParameter)
	withInfo = viper.GetBool(withInfoParameter)

	msglog.SetLogLevelFromString(viper.GetString(logLevelParameter))
}
