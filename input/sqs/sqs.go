package sqs

import (
	"context"
	"fmt"
	"log"

	"github.com/docktermj/go-xyzzy-helpers/logger"
	"github.com/roncewind/move/io/rabbitmq/managedconsumer"
	"github.com/roncewind/move/io/sqs"

	"github.com/senzing/g2-sdk-go/g2api"
	"github.com/senzing/go-sdk-abstract-factory/factory"
)

// load is 6201:  https://github.com/Senzing/knowledge-base/blob/main/lists/senzing-product-ids.md
const MessageIdFormat = "senzing-6201%04d"

// ----------------------------------------------------------------------------

// read and process records from the given queue until a system interrupt
func Read(ctx context.Context, urlString, engineConfigJson string, engineLogLevel int, numberOfWorkers int) {

	// Work with G2engine.
	g2engine := createG2Engine(ctx, engineConfigJson, engineLogLevel)
	defer g2engine.Destroy(ctx)
	// var g2engine *g2api.G2engine = nil

	// fmt.Println(" [*] Waiting for messages. To exit press CTRL+C")
	fmt.Println("reading:", urlString)
	startErr := sqs.StartConsumer(ctx, urlString, numberOfWorkers, g2engine, false, 60)
	if startErr != nil {
		msg := "there was an unexpected issue; please report this as a bug."
		if _, ok := startErr.(managedconsumer.ManagedConsumerError); ok {
			msg = startErr.Error()
		}
		handleError(1, startErr, msg)
	}
	fmt.Println("So long and thanks for all the fish.")
}

// ----------------------------------------------------------------------------

// create a G2Engine object, on error this function panics.
// see failOnError
func createG2Engine(ctx context.Context, engineConfigJson string, engineLogLevel int) g2api.G2engine {
	senzingFactory := &factory.SdkAbstractFactoryImpl{}
	g2Config, err := senzingFactory.GetG2config(ctx)
	if err != nil {
		handleError(2, err, "Unable to retrieve the config")
	}
	g2engine, err := senzingFactory.GetG2engine(ctx)

	if err != nil {
		logger.LogMessage(MessageIdFormat, 2000, err.Error())
		handleError(3, err, "Unable to reach G2")
	}
	if g2Config.GetSdkId(ctx) == "base" {
		err = g2engine.Init(ctx, "load", engineConfigJson, engineLogLevel)
		if err != nil {
			handleError(4, err, "Could not Init G2")
		}
	}
	return g2engine
}

// ----------------------------------------------------------------------------

// TODO: update error handling
func handleError(key int, err error, msg string) {
	log.SetPrefix(fmt.Sprintf("[logID:%v]", key))
	log.Printf("%#v\n", err)
	fmt.Printf("[%v] %v", key, msg)
	panic(err)
}
