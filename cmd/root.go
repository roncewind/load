/*
Copyright © 2022 roncewind <dad@lynntribe.net>
*/
package cmd

import (
	// "encoding/json"
	// "errors"
	"fmt"
	"strings"

	// "net"
	// "net/http"
	// "net/url"
	"os"

	"github.com/docktermj/go-xyzzy-helpers/logger"
	"github.com/roncewind/load/input"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile    string
	exchange   string = "senzing"
	fileType   string //TODO: load from file
	inputQueue string = "senzing_input"
	inputURL   string // read from this URL, could be a file or a queue
	logLevel   string = "error"
	withInfo   bool   = false
)

// load is 6201:  https://github.com/Senzing/knowledge-base/blob/main/lists/senzing-product-ids.md
const MessageIdFormat = "senzing-6201%04d"

// ----------------------------------------------------------------------------
// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "load",
	Short: "Load records into Senzing",
	Long: `TODO: A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("start load Run")
		fmt.Println("viper key list:")
		for _, key := range viper.AllKeys() {
			fmt.Println("  - ", key, " = ", viper.Get(key))
		}

		if !input.Read() {
			cmd.Help()
		}

	},
}

// ----------------------------------------------------------------------------
// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	fmt.Println("start load Execute")
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// ----------------------------------------------------------------------------
func init() {
	fmt.Println("start load init")
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.senzing/config.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().StringVarP(&exchange, "exchange", "", "", "Message queue exchange name")
	viper.BindPFlag("exchange", RootCmd.Flags().Lookup("exchange"))
	RootCmd.Flags().StringVarP(&fileType, "fileType", "", "", "file type override")
	viper.BindPFlag("fileType", RootCmd.Flags().Lookup("fileType"))
	RootCmd.Flags().StringVarP(&inputQueue, "inputQueue", "", "", "Senzing input queue name")
	viper.BindPFlag("inputQueue", RootCmd.Flags().Lookup("inputQueue"))
	RootCmd.Flags().StringVarP(&inputURL, "inputURL", "i", "", "input location")
	viper.BindPFlag("inputURL", RootCmd.Flags().Lookup("inputURL"))
	RootCmd.Flags().StringVarP(&logLevel, "logLevel", "", "", "set the logging level, default Error")
	viper.BindPFlag("logLevel", RootCmd.Flags().Lookup("logLevel"))
	RootCmd.Flags().BoolP("withInfo", "", false, "set to add record withInfo")
	viper.BindPFlag("withInfo", RootCmd.Flags().Lookup("withInfo"))
}

// ----------------------------------------------------------------------------
// initConfig reads in config file and ENV variables if set.
// Config precedence:
// - cmdline args
// - env vars
// - config file
func initConfig() {
	fmt.Println("start load initConfig")
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
			logger.LogMessageFromError(MessageIdFormat, 2001, "Config file found, but not loaded", err)
		}
	}
	viper.AutomaticEnv() // read in environment variables that match
	// all env vars should be prefixed with "SENZING_TOOLS_"
	viper.SetEnvPrefix("senzing_tools")
	viper.BindEnv("exchange")
	viper.BindEnv("fileType")
	viper.BindEnv("inputQueue")
	viper.BindEnv("inputURL")
	viper.BindEnv("logLevel")
	viper.BindEnv("withInfo")

	fmt.Printf("var-->>%s\n", inputURL)
	fmt.Printf("from viper-->>%s\n", viper.GetString("inputURL"))
	viper.SetDefault("exchange", "senzing")
	viper.SetDefault("inputQueue", "senzing-input")
	viper.SetDefault("logLevel", "error")
	viper.SetDefault("withInfo", false)

	// setup local variables, in case they came from a config file
	//TODO:  why do I have to do this?  env vars and cmdline params get mapped
	//  automatically, this is only IF the var is in the config file
	//FIXME:  this over writes cmdline args when used from senzing-tools

	// exchange = viper.GetString("exchange")
	// fileType = viper.GetString("fileType")
	// inputQueue = viper.GetString("inputQueue")
	// inputURL = viper.GetString("inputURL")
	// logLevel = viper.GetString("logLevel")
	// withInfo = viper.GetBool("withInfo")
	// fmt.Printf("2-->>%s\n", inputURL)
	setLogLevel()
}

// ----------------------------------------------------------------------------
func setLogLevel() {
	var level logger.Level = logger.LevelError
	if viper.IsSet("logLevel") {
		switch strings.ToUpper(logLevel) {
		case logger.LevelDebugName:
			level = logger.LevelDebug
		case logger.LevelErrorName:
			level = logger.LevelError
		case logger.LevelFatalName:
			level = logger.LevelFatal
		case logger.LevelInfoName:
			level = logger.LevelInfo
		case logger.LevelPanicName:
			level = logger.LevelPanic
		case logger.LevelTraceName:
			level = logger.LevelTrace
		case logger.LevelWarnName:
			level = logger.LevelWarn
		}
		logger.SetLevel(level)
	}
}
