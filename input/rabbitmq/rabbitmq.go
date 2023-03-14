package rabbitmq

import (
	"context"
	"fmt"
	"os"

	"github.com/docktermj/go-xyzzy-helpers/logger"
	"github.com/roncewind/move/io/rabbitmq/managedconsumer"

	"github.com/senzing/g2-sdk-go/g2api"
	"github.com/senzing/go-sdk-abstract-factory/factory"
)

// load is 6201:  https://github.com/Senzing/knowledge-base/blob/main/lists/senzing-product-ids.md
const MessageIdFormat = "senzing-6201%04d"

// ----------------------------------------------------------------------------
// TODO: rename?
func Read(ctx context.Context, urlString string) {

	// Work with G2engine.
	g2engine := createG2Engine(ctx)
	defer (*g2engine).Destroy(ctx)

	// fmt.Println(" [*] Waiting for messages. To exit press CTRL+C")
	fmt.Println("reading:", urlString)
	<-managedconsumer.StartManagedConsumer(ctx, urlString, 0, g2engine, false)

}

// ----------------------------------------------------------------------------

// create a G2Engine object, on error this function panics.
// see failOnError
func createG2Engine(ctx context.Context) *g2api.G2engine {
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
	}
	return &g2engine
}

// ----------------------------------------------------------------------------

// TODO: update error handling
func failOnError(err error, msg string) {
	if err != nil {
		s := fmt.Sprintf("%s: %s", msg, err)
		panic(s)
	}
}
