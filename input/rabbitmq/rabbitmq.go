package rabbitmq

import (
	// "encoding/json"
	// "errors"
	"fmt"
	// "net"
	// "net/http"
	// "net/url"
	// "os"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/roncewind/szrecord"
	// "github.com/spf13/cobra"
	// "github.com/spf13/viper"
)

// ----------------------------------------------------------------------------
// TODO: move to a new module in order to reuse
func Read(urlString string, exchange string, queue string) {
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
