/*
Copyright © 2022 roncewind <dad@lynntribe.net>
*/
package cmd

import (
	"os"

	"github.com/roncewind/load/input"
	"github.com/senzing/go-logging/logger"
	"github.com/senzing/go-logging/messagelogger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile    string
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

load --inputURL "amqp://guest:guest@192.168.6.128:5672
`,

	Run: func(cmd *cobra.Command, args []string) {
		if msglog.IsInfo() {
			msglog.Log(1, logger.LevelInfo)
			for _, key := range viper.AllKeys() {
				msglog.Log(2, key, viper.Get(key), logger.LevelInfo)
			}
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
	RootCmd.Flags().StringVarP(&fileType, "fileType", "", "", "file type override")
	viper.BindPFlag("fileType", RootCmd.Flags().Lookup("fileType"))
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
	viper.SetEnvPrefix("senzing_tools")
	viper.BindEnv("fileType")
	viper.BindEnv("inputURL")
	viper.BindEnv("logLevel")
	viper.BindEnv("withInfo")

	// cmdline args should get set in viper, but for some reason that's
	// not happening when called from senzing-tools, this is the work around:
	if len(fileType) > 0 {
		viper.Set("fileType", fileType)
	}
	if len(inputURL) > 0 {
		viper.Set("inputURL", inputURL)
	}
	if len(logLevel) > 0 {
		viper.Set("logLevel", logLevel)
	}
	if withInfo {
		viper.Set("withInfo", withInfo)
	}

	viper.SetDefault("logLevel", "info")
	viper.SetDefault("withInfo", false)

	// setup local variables, in case they came from a config file
	//TODO:  why do I have to do this?  env vars and cmdline params get mapped
	//  automatically, this is only IF the var is in the config file.
	//  am i missing a way to bind config file vars to local vars?
	fileType = viper.GetString("fileType")
	inputURL = viper.GetString("inputURL")
	logLevel = viper.GetString("logLevel")
	withInfo = viper.GetBool("withInfo")

	msglog.SetLogLevelFromString(logLevel)
}
