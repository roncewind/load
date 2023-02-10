package rabbitmq

import (
	"context"
	"fmt"

	"github.com/docktermj/go-xyzzy-helpers/logger"
	"github.com/roncewind/move/io/rabbitmq/managedconsumer"
	"github.com/senzing/g2-sdk-go/g2engine"
	"github.com/senzing/go-helpers/g2engineconfigurationjson"
)

// load is 6201:  https://github.com/Senzing/knowledge-base/blob/main/lists/senzing-product-ids.md
const MessageIdFormat = "senzing-6201%04d"

// ----------------------------------------------------------------------------
// TODO: rename
func Read(urlString string, exchangeName string, queueName string) {

	ctx := context.Background()

	// Work with G2engine.

	g2engine, g2engineErr := getG2engine(ctx)
	defer g2engine.Destroy(ctx)

	if g2engineErr != nil {
		logger.LogMessage(MessageIdFormat, 2000, g2engineErr.Error())
		failOnError(g2engineErr, "Unable to reach G2")
	}

	// fmt.Println(" [*] Waiting for messages. To exit press CTRL+C")
	<-managedconsumer.StartManagedConsumer(ctx, exchangeName, queueName, urlString, 3, g2engine, false)

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
