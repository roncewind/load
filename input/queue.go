package input

import (
	// "encoding/json"
	// "errors"
	"fmt"
	"net"
	// "net/http"
	"net/url"
	// "os"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/roncewind/szrecord"
	// "github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ----------------------------------------------------------------------------
func parseURL(urlString string) (*url.URL){
	u, err := url.Parse(urlString)
	if err != nil {
		panic(err)
	}


	fmt.Println("===============================")
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
	fmt.Println("===============================")

	return u
}

// ----------------------------------------------------------------------------
func Read() (bool) {
	u := parseURL(viper.GetString("inputURL"))
	switch u.Scheme {
	case "amqp":
		if( viper.IsSet("inputURL") &&
			viper.IsSet("exchange") &&
			viper.IsSet("inputQueue")) {
			ReadRabbit(viper.GetString("inputURL"), viper.GetString("exchange"), viper.GetString("inputQueue"))
		} else {
			return false
		}
	default:
		fmt.Println("Unknown input mechanism: %s", u.Scheme)
	}
	return true
}

// ----------------------------------------------------------------------------
// TODO: move to a new module in order to reuse
func ReadRabbit(urlString string, exchange string, queue string) {
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
			valid, err := szrecord.Validate(string(d.Body))
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
func failOnError(err error, msg string) {
	if err != nil {
		s := fmt.Sprintf("%s: %s", msg, err)
		panic(s)
	}
}
