package rabbitmq

import (
	"context"
	"fmt"

	"github.com/docktermj/go-xyzzy-helpers/logger"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/roncewind/move/io/rabbitmq/managedconsumer"
	"github.com/roncewind/szrecord"
	"github.com/senzing/g2-sdk-go/g2engine"
	"github.com/senzing/go-helpers/g2engineconfigurationjson"
	"github.com/spf13/viper"
)

// load is 6201:  https://github.com/Senzing/knowledge-base/blob/main/lists/senzing-product-ids.md
const MessageIdFormat = "senzing-6201%04d"

// ----------------------------------------------------------------------------
// TODO: rename
func Read(urlString string, exchangeName string, queueName string) {

	ctx := context.TODO()

	// Work with G2engine.

	g2engine, g2engineErr := getG2engine(ctx)
	defer g2engine.Destroy(ctx)

	if g2engineErr != nil {
		logger.LogMessage(MessageIdFormat, 2000, g2engineErr.Error())
		failOnError(g2engineErr, "Unable to reach G2")
	}

	// fmt.Println(" [*] Waiting for messages. To exit press CTRL+C")
	managedconsumer.StartManagedConsumer(exchangeName, queueName, urlString, 3, g2engine, false)

}

// ----------------------------------------------------------------------------
// TODO: rename
func ReadXXX(urlString string, exchange string, queue string) {

	ctx := context.TODO()

	// Work with G2engine.

	g2engine, g2engineErr := getG2engine(ctx)
	if g2engineErr != nil {
		logger.LogMessage(MessageIdFormat, 2000, g2engineErr.Error())
		failOnError(g2engineErr, "Unable to reach G2")
	}

	conn, err := amqp.Dial(urlString)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		exchange, // name
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
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name,   // queue name
		q.Name,   // routing key
		exchange, // exchange
		false,
		nil,
	)

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
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

	//TODO: input param for number of go routines
	for i := 1; i < 3; i++ {
		go handler(ctx, g2engine, msgs)
	}

	fmt.Println(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

// ----------------------------------------------------------------------------
func handler(ctx context.Context, g2engine g2engine.G2engine, msgs <-chan amqp.Delivery) {
	for d := range msgs {
		fmt.Printf("Received a message- msgId: %s, msgCnt: %d, ConsumerTag: %s\n", d.MessageId, d.MessageCount, d.ConsumerTag)
		record, newRecordErr := szrecord.NewRecord(string(d.Body))
		if newRecordErr == nil {

			fmt.Printf("Processing record: %s\n", record.Id)
			loadID := "Load"
			if viper.GetBool("withInfo") {
				var flags int64 = 0
				withInfo, withInfoErr := g2engine.AddRecordWithInfo(ctx, record.DataSource, record.Id, record.Json, loadID, flags)
				if withInfoErr != nil {
					logger.LogMessage(MessageIdFormat, 2002, withInfoErr.Error())
				} else {
					//TODO:  what do we do with the record here?
					fmt.Printf("Record[%s] Added: %s:%s:%s\n", d.MessageId, loadID, record.DataSource, record.Id)
					fmt.Printf("WithInfo: %s\n", withInfo)
				}
			} else {
				addRecordErr := g2engine.AddRecord(ctx, record.DataSource, record.Id, record.Json, loadID)
				if addRecordErr != nil {
					logger.LogMessage(MessageIdFormat, 2003, addRecordErr.Error())
				} else {
					fmt.Printf("Record[%s] Added: %s:%s:%s\n", d.MessageId, loadID, record.DataSource, record.Id)
				}

			}

			// when we successfully process a delivery, Ack it.
			d.Ack(false)
			// when there's an issue with a delivery should we requeue it?
			// d.Nack(false, true)
		} else {
			logger.LogMessageFromError(MessageIdFormat, 2001, "create new szRecord", newRecordErr)
			// when we get an invalid delivery, Ack it, so we don't requeue
			// TODO: set up rabbit with a dead letter queue?
			d.Ack(false)
			// FIXME: errors should be specific to the input method
			//  ala rabbitmq message ID?
		}
	}
}

// ----------------------------------------------------------------------------
func getG2engine(ctx context.Context) (g2engine.G2engine, error) {
	var err error = nil
	g2engine := g2engine.G2engineImpl{}

	moduleName := "Load"
	verboseLogging := 0 // 0 for no Senzing logging; 1 for logging
	iniParams, jsonErr := g2engineconfigurationjson.BuildSimpleSystemConfigurationJson("")
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
