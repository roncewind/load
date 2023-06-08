/*
Copyright © 2022 roncewind <dad@lynntribe.net>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/docktermj/go-xyzzy-helpers/logger"
	"github.com/roncewind/load/loader"
	"github.com/senzing/senzing-tools/constant"
	"github.com/senzing/senzing-tools/envar"
	"github.com/senzing/senzing-tools/helper"
	"github.com/senzing/senzing-tools/option"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	defaultDelayInSeconds            int    = 0
	defaultEngineConfig              string = ""
	defaultEngineLogLevel            int    = 0
	defaultFileType                  string = ""
	defaultInputURL                  string = ""
	defaultOutputURL                 string = ""
	defaultLogLevel                  string = "error"
	defaultNumberOfWorkers           int    = 0
	defaultVisibilityPeriodInSeconds int    = 60
	Use                              string = "load"
	Short                            string = "Load records into Senzing."
	Long                             string = `
	Welcome to load!
	This tool will load records into Senzing. It validates the records conform to the Generic Entity Specification.

	For example:

	load --input-url "amqp://guest:guest@192.168.6.96:5672"
	load --input-url "https://public-read-access.s3.amazonaws.com/TestDataSets/SenzingTruthSet/truth-set-3.0.0.jsonl"
`
)

const (
	envVarReplacerCharNew string = "_"
	envVarReplacerCharOld string = "-"
	withInfoParameter     string = "with-info"
)

var (
	buildIteration string = "0"
	buildVersion   string = "0.0.0"
	programName    string = "load"

	defaultWithInfoParameter string = ""
)

// load is 6201:  https://github.com/Senzing/knowledge-base/blob/main/lists/senzing-product-ids.md
const productIdentifier = 6201

// ----------------------------------------------------------------------------
// Command
// ----------------------------------------------------------------------------

// RootCmd represents the command.
var RootCmd = &cobra.Command{
	Use:     Use,
	Short:   Short,
	Long:    Long,
	PreRun:  PreRun,
	Run:     Run,
	Version: Version(),
}

// ----------------------------------------------------------------------------

// Used in construction of cobra.Command
func PreRun(cobraCommand *cobra.Command, args []string) {
	loadConfigurationFile(cobraCommand)
	loadOptions(cobraCommand)
	cobraCommand.SetVersionTemplate(constant.VersionTemplate)
}

// ----------------------------------------------------------------------------

// The core of this command
func Run(cmd *cobra.Command, args []string) {
	fmt.Println("Run with the following parameters:")
	for _, key := range viper.AllKeys() {
		fmt.Println("  - ", key, " = ", viper.Get(key))
	}
	setLogLevel()
	if viper.GetInt(option.DelayInSeconds) > 0 {
		fmt.Println(time.Now(), "Sleep for", viper.GetInt(option.DelayInSeconds), "seconds to let queues and databases settle down and come up.")
		time.Sleep(time.Duration(viper.GetInt(option.DelayInSeconds)) * time.Second)
	}

	if viper.IsSet(option.InputURL) {
		ctx := context.Background()

		loader := &loader.LoaderImpl{
			EngineConfigJson:          viper.GetString(option.EngineConfigurationJson),
			EngineLogLevel:            viper.GetInt(option.EngineLogLevel),
			InputURL:                  viper.GetString(option.InputURL),
			LogLevel:                  viper.GetString(option.LogLevel),
			NumberOfWorkers:           viper.GetInt(option.NumberOfWorkers),
			VisibilityPeriodInSeconds: viper.GetInt(option.VisibilityPeriodInSeconds),
		}

		if !loader.Load(ctx) {
			cmd.Help()
		}
	} else {
		cmd.Help()
		fmt.Println("Build Version:", buildVersion)
		fmt.Println("Build Iteration:", buildIteration)
	}

}

// ----------------------------------------------------------------------------

// Used in construction of cobra.Command
func Version() string {
	return helper.MakeVersion(githubVersion, githubIteration)
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
	RootCmd.Flags().Int(option.DelayInSeconds, defaultDelayInSeconds, option.DelayInSecondsHelp)
	RootCmd.Flags().Int(option.EngineLogLevel, defaultEngineLogLevel, option.EngineLogLevel)
	RootCmd.Flags().String(option.InputFileType, defaultFileType, option.InputFileTypeHelp)
	RootCmd.Flags().String(option.InputURL, defaultInputURL, option.InputURLHelp)
	RootCmd.Flags().String(option.LogLevel, defaultLogLevel, fmt.Sprintf(option.LogLevelHelp, envar.LogLevel))
	RootCmd.Flags().Int(option.NumberOfWorkers, defaultNumberOfWorkers, option.NumberOfWorkersHelp)
	RootCmd.Flags().Int(option.VisibilityPeriodInSeconds, defaultVisibilityPeriodInSeconds, option.VisibilityPeriodInSecondsHelp)
	RootCmd.Flags().String(option.OutputURL, defaultOutputURL, option.OutputURLHelp)
	runtime.GOMAXPROCS(64)
}

// ----------------------------------------------------------------------------

// If a configuration file is present, load it.
func loadConfigurationFile(cobraCommand *cobra.Command) {
	configuration := ""
	configFlag := cobraCommand.Flags().Lookup(option.Configuration)
	if configFlag != nil {
		configuration = configFlag.Value.String()
	}
	if configuration != "" { // Use configuration file specified as a command line option.
		viper.SetConfigFile(configuration)
	} else { // Search for a configuration file.

		// Determine home directory.

		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Specify configuration file name.

		viper.SetConfigName("move")
		viper.SetConfigType("yaml")

		// Define search path order.

		viper.AddConfigPath(home + "/.senzing-tools")
		viper.AddConfigPath(home)
		viper.AddConfigPath("/etc/senzing-tools")
	}

	// If a config file is found, read it in.

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Applying configuration file:", viper.ConfigFileUsed())
	}
}

// ----------------------------------------------------------------------------

// Configure Viper with user-specified options.
func loadOptions(cobraCommand *cobra.Command) {
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(envVarReplacerCharOld, envVarReplacerCharNew)
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix(constant.SetEnvPrefix)

	// Ints

	intOptions := map[string]int{
		option.DelayInSeconds:            defaultDelayInSeconds,
		option.EngineLogLevel:            defaultEngineLogLevel,
		option.NumberOfWorkers:           defaultNumberOfWorkers,
		option.VisibilityPeriodInSeconds: defaultVisibilityPeriodInSeconds,
	}
	for optionKey, optionValue := range intOptions {
		viper.SetDefault(optionKey, optionValue)
		viper.BindPFlag(optionKey, cobraCommand.Flags().Lookup(optionKey))
	}

	// Strings

	stringOptions := map[string]string{
		option.EngineConfigurationJson: defaultEngineConfig,
		option.InputFileType:           defaultFileType,
		option.InputURL:                defaultInputURL,
		option.LogLevel:                defaultLogLevel,
		withInfoParameter:              defaultWithInfoParameter,
	}
	for optionKey, optionValue := range stringOptions {
		viper.SetDefault(optionKey, optionValue)
		viper.BindPFlag(optionKey, cobraCommand.Flags().Lookup(optionKey))
	}

}

// ----------------------------------------------------------------------------
func setLogLevel() {
	var level logger.Level = logger.LevelError
	if viper.IsSet(option.LogLevel) {
		switch strings.ToUpper(viper.GetString(option.LogLevel)) {
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
