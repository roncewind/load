package rabbitmq

import (
	"context"
	"fmt"
	"os"

	"github.com/docktermj/go-xyzzy-helpers/logger"
	"github.com/roncewind/move/io/rabbitmq/managedconsumer"

	"github.com/senzing/go-sdk-abstract-factory/factory"
)

// load is 6201:  https://github.com/Senzing/knowledge-base/blob/main/lists/senzing-product-ids.md
const MessageIdFormat = "senzing-6201%04d"

// ----------------------------------------------------------------------------
// TODO: rename?
func Read(urlString string) {

	ctx := context.Background()

	// Work with G2engine.
	senzingFactory := &factory.SdkAbstractFactoryImpl{}
	g2Config, err := senzingFactory.GetG2config(ctx)
	if err != nil {
		failOnError(err, "Unable to retrieve the config")
	}
	g2engine, err := senzingFactory.GetG2engine(ctx)

	if err != nil {
		logger.LogMessage(MessageIdFormat, 2000, err.Error())
		failOnError(err, "Unable to reach G2")
	}
	if g2Config.GetSdkId(ctx) == "base" {
		configJSON, _ := os.LookupEnv("SENZING_ENGINE_CONFIGURATION_JSON")
		err = g2engine.Init(ctx, "load", configJSON, 0)
		if err != nil {
			failOnError(err, "Could not Init G2")
		}
		defer g2engine.Destroy(ctx)
	}

	// fmt.Println(" [*] Waiting for messages. To exit press CTRL+C")
	fmt.Println("reading:", urlString)
	<-managedconsumer.StartManagedConsumer(ctx, urlString, 0, g2engine, false)

}

// ----------------------------------------------------------------------------

// TODO: update error handling
func failOnError(err error, msg string) {
	if err != nil {
		s := fmt.Sprintf("%s: %s", msg, err)
		panic(s)
	}
}
