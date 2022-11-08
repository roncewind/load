/*
Copyright © 2022 roncewind <dad@lynntribe.net>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/roncewind/load/input"
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
	logLevel   string = "warn"
	logger     messagelogger.MessageLoggerInterface
	withInfo   bool = false
)

// load is 6201:  https://github.com/Senzing/knowledge-base/blob/main/lists/senzing-product-ids.md
const productIdentifier = 6201

var idMessages = map[int]string{
	0:    "Logger initialized.",
	5:    "The favorite number for %s is %d.",
	6:    "Person number #%[2]d is %[1]s.",
	10:   "Example errors.",
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
	logger, _ = messagelogger.NewSenzingLogger(productIdentifier, idMessages)
	logger.Log(0)
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.senzing/config.yaml)")

	// local flags for load
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
	fmt.Printf("logger: %v\n", logger)
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
			logger.Log(2001, err)
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

	// cmdline args should get set in viper, but for some reason that's
	// not happening when called from senzing-tools, this is the work around:
	if len(exchange) > 0 {
		viper.Set("exchange", exchange)
	}
	if len(fileType) > 0 {
		viper.Set("fileType", fileType)
	}
	if len(inputQueue) > 0 {
		viper.Set("inputQueue", inputQueue)
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

	viper.SetDefault("exchange", "senzing")
	viper.SetDefault("inputQueue", "senzing-input")
	viper.SetDefault("logLevel", "error")
	viper.SetDefault("withInfo", false)

	// setup local variables, in case they came from a config file
	//TODO:  why do I have to do this?  env vars and cmdline params get mapped
	//  automatically, this is only IF the var is in the config file.
	//  am i missing a way to bind config file vars to local vars?
	exchange = viper.GetString("exchange")
	fileType = viper.GetString("fileType")
	inputQueue = viper.GetString("inputQueue")
	inputURL = viper.GetString("inputURL")
	logLevel = viper.GetString("logLevel")
	withInfo = viper.GetBool("withInfo")

	logger.SetLogLevelFromString(logLevel)
}
