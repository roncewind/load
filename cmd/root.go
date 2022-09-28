/*
Copyright © 2022 roncewind <dad@lynntribe.net>

*/
package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	// "net/http"
	"net/url"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string

	exchange string
	inputQueue string

	inputURL  string // read from this URL, could be a file or a queue
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "load",
	Short: "Load records into Senzing",
	Long: `TODO: A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("start Run")
		fmt.Println("viper key list:")
		for _, key := range viper.AllKeys() {
			fmt.Println("  - ", key, " = ", viper.Get(key))
		}

		//TODO: test for required parameters otherwise show help.
		if( viper.IsSet("inputURL")
		&& viper.IsSet("exchange") && viper.IsSet("inputQueue")) {
			parseURL(viper.GetString("inputURL"))
			read(viper.GetString("inputURL"), viper.GetString("exchange"), viper.GetString("inputQueue"))
		} else {
			cmd.Help()
		}

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.senzing/config.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	RootCmd.Flags().StringVarP(&inputURL, "inputURL", "i", "", "input location")
	viper.BindPFlag("inputURL", RootCmd.Flags().Lookup("inputURL"))
	RootCmd.Flags().StringVarP(&exchange, "exchange", "", "", "Message queue exchange name")
	viper.BindPFlag("exchange", RootCmd.Flags().Lookup("exchange"))
	RootCmd.Flags().StringVarP(&inputQueue, "inputQueue", "", "", "Senzing input queue name")
	viper.BindPFlag("inputQueue", RootCmd.Flags().Lookup("inputQueue"))

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in <home directory>/.senzing with name "config" (without extension).
		viper.AddConfigPath(home+"/.senzing-tools")
		viper.AddConfigPath(home)
		viper.AddConfigPath("/etc/senzing-tools")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match
	// all env vars should be prefixed with "SENZING_TOOLS_"
	viper.SetEnvPrefix("senzing_tools")
	viper.BindEnv("inputURL")
	viper.BindEnv("exchange")
	viper.SetDefault("exchange", "senzing")
	viper.BindEnv("inputQueue")
	viper.SetDefault("inputQueue", "senzing-input")

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func parseURL(urlString string) {
	u, err := url.Parse(urlString)
	if err != nil {
		panic(err)
	}

	fmt.Println("\tScheme: ", u.Scheme)
	fmt.Println("\tUser full: ", u.User)
	fmt.Println("\tUser name: ", u.User.Username())
	p, _ := u.User.Password()
	fmt.Println("\tPassword: ", p)

	fmt.Println("\tHost full: ", u.Host)
	host, port, _ := net.SplitHostPort(u.Host)
	fmt.Println("\tHost: ", host)
	fmt.Println("\tPort: ", port)

	fmt.Println("\tPath: ", u.Path)
	fmt.Println("\tFragment: ", u.Fragment)

	fmt.Println("\tQuery string: ", u.RawQuery)
	m, _ := url.ParseQuery(u.RawQuery)
	fmt.Println("\tParsed query string: ", m)
	// fmt.Println(m["k"][0])
}


func failOnError(err error, msg string) {
	if err != nil {
		s := fmt.Sprintf("%s: %s", msg, err)
		panic(s)
	}
}

func read(urlString string, exchange string, queue string) {
	conn, err := amqp.Dial(urlString)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		exchange,   // name
		"direct", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	q, err := ch.QueueDeclare(
	  queue, // name
	  true,   // durable
	  false,   // delete when unused
	  false,   // exclusive
	  false,   // no-wait
	  nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name, // queue name
		q.Name,     // routing key
		exchange, // exchange
		false,
		nil,
	)

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a loader")

	err = ch.Qos(
		3,     // prefetch count (should set to the number of load goroutines)
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			fmt.Printf("Received a message: %s\n", d.Body)
			valid, err := validateLine(string(d.Body))
			if valid {
				//TODO: Senzing here
				// when we successfully process a delivery, Ack it.
				d.Ack(false)
				// when there's an issue with a delivery should we requeue it?
				// d.Nack(false, true)
			} else {
				// when we get an invalid delivery, Ack it, so we don't requeue
				d.Ack(false)
				// FIXME: errors should be specific to the input method
				//  ala rabbitmq message ID?
				fmt.Println("Error with message:", err)
			}
		}
	}()

	fmt.Println(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

// ----------------------------------------------------------------------------
type Record struct {
	DataSource string `json:"DATA_SOURCE"`
	RecordId string `json:"RECORD_ID"`
}

// ----------------------------------------------------------------------------
func validateLine(line string) (bool, error) {
	var record Record
	valid := json.Unmarshal([]byte(line), &record) == nil
	if valid {
		return validateRecord(record)
	}
	return valid, errors.New("JSON-line not well formed.")
}

// ----------------------------------------------------------------------------
func validateRecord(record Record) (bool, error) {
	// FIXME: errors should be specific to the input method
	//  ala rabbitmq message ID?
	if record.DataSource == "" {
		return false, errors.New("A DATA_SOURCE field is required.")
	}
	if record.RecordId == "" {
		return false, errors.New("A RECORD_ID field is required.")
	}
	return true, nil
}
