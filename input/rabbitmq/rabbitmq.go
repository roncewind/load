package rabbitmq

import (
	"context"
	// "encoding/json"
	// "errors"
	"fmt"
	// "log"
	// "net"
	// "net/http"
	// "net/url"
	// "os"

	"github.com/docktermj/g2-sdk-go/g2engine"
	"github.com/docktermj/go-xyzzy-helpers/g2configuration"
	"github.com/docktermj/go-xyzzy-helpers/logger"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/roncewind/szrecord"
	// "github.com/spf13/cobra"
	// "github.com/spf13/viper"
)

const MessageIdFormat = "senzing-LOAD%04d"

// ----------------------------------------------------------------------------
// TODO: rename
func Read(urlString string, exchange string, queue string) {

	ctx := context.TODO()

	// Work with G2engine.

	g2engine, g2engineErr := getG2engine(ctx)
	if g2engineErr != nil {
		logger.LogMessage(MessageIdFormat, 5, g2engineErr.Error())
		failOnError(g2engineErr, "Unable to reach G2")
	}

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
			valid := false
			record, err := szrecord.NewRecord(string(d.Body))
			if err == nil {
				isValid, err := szrecord.ValidateRecord(*record)
				valid = isValid
			} else {
				logger.LogMessageFromError(MessageIdFormat, 5, "", err)
				fmt.Println("Validation error message:", err)
			}
			if valid {
				//TODO: Senzing here
				loadID := "Load"
				var flags int64 = 0

				withInfo, withInfoErr := g2engine.AddRecordWithInfo(ctx, record.DataSource, record.Id, record.Json, loadID, flags)
				if withInfoErr != nil {
					logger.LogMessage(MessageIdFormat, 5, withInfoErr.Error())
					fmt.Println("Withinfo error message:", withInfoErr)
				}

				fmt.Printf("WithInfo: %s\n", withInfo)
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
func getG2engine(ctx context.Context) (g2engine.G2engine, error) {
	var err error = nil
	g2engine := g2engine.G2engineImpl{}

	moduleName := "Load"
	verboseLogging := 0 // 0 for no Senzing logging; 1 for logging
	iniParams, jsonErr := g2configuration.BuildSimpleSystemConfigurationJson("")
	if jsonErr != nil {
		return &g2engine, jsonErr
	}

	err = g2engine.Init(ctx, moduleName, iniParams, verboseLogging)
	return &g2engine, err
}

// ----------------------------------------------------------------------------
func failOnError(err error, msg string) {
	if err != nil {
		s := fmt.Sprintf("%s: %s", msg, err)
		panic(s)
	}
}
