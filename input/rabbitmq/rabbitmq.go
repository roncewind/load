package rabbitmq

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"time"

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

	logBuildInfo()
	logStats()
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				logStats()
			}
		}
	}()
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

// ----------------------------------------------------------------------------

func logBuildInfo() {
	buildInfo, ok := debug.ReadBuildInfo()
	fmt.Println("---------------------------------------------------------------")
	if ok {
		fmt.Println("GoVersion:", buildInfo.GoVersion)
		fmt.Println("Path:", buildInfo.Path)
		fmt.Println("Main.Path:", buildInfo.Main.Path)
		fmt.Println("Main.Version:", buildInfo.Main.Version)
	} else {
		fmt.Println("Unable to read build info.")
	}
}

// ----------------------------------------------------------------------------

func logStats() {
	cpus := runtime.NumCPU()
	goRoutines := runtime.NumGoroutine()
	cgoCalls := runtime.NumCgoCall()
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	var gcStats debug.GCStats
	debug.ReadGCStats(&gcStats)

	// fmt.Println("---------------------------------------------------------------")
	// fmt.Println("Time:", time.Now())
	// fmt.Println("CPUs:", cpus)
	// fmt.Println("Go routines:", goRoutines)
	// fmt.Println("CGO calls:", cgoCalls)
	// fmt.Println("Num GC:", memStats.NumGC)
	// fmt.Println("GCSys:", memStats.GCSys)
	// fmt.Println("GC pause total:", gcStats.PauseTotal)
	// fmt.Println("LastGC:", gcStats.LastGC)
	// fmt.Println("HeapAlloc:", memStats.HeapAlloc)
	// fmt.Println("NextGC:", memStats.NextGC)
	// fmt.Println("CPU fraction used by GC:", memStats.GCCPUFraction)

	fmt.Println("---------------------------------------------------------------")
	printCSV(">>>", "Time", "CPUs", "Go routines", "CGO calls", "Num GC", "GC pause total", "LastGC", "TotalAlloc", "HeapAlloc", "NextGC", "GCSys", "HeapSys", "StackSys", "Sys - total OS bytes", "CPU fraction used by GC")
	printCSV(">>>", time.Now(), cpus, goRoutines, cgoCalls, memStats.NumGC, gcStats.PauseTotal, gcStats.LastGC, memStats.TotalAlloc, memStats.HeapAlloc, memStats.NextGC, memStats.GCSys, memStats.HeapSys, memStats.StackSys, memStats.Sys, memStats.GCCPUFraction)
}

// ----------------------------------------------------------------------------

func printCSV(fields ...any) {
	for _, field := range fields {
		fmt.Print(field, ",")
	}
	fmt.Println("")
}
