package main

import (
	"context"
	"log"

	"github.com/BuxOrg/bux"
	"github.com/BuxOrg/bux/taskmanager"
)

func main() {
	client, err := bux.NewClient(
		context.Background(), // Set context
		bux.WithTaskQ(taskmanager.DefaultTaskQConfig("test_queue"), taskmanager.FactoryMemory), // Tasks
		bux.WithDebugging(), // Enable debugging (verbose logs)
		bux.WithChainstateOptions(true, true, true, true), // Broadcasting enabled by default
	)
	if err != nil {
		log.Fatalln("error: " + err.Error())
	}

	log.Println("client loaded!", client.UserAgent(), "debugging: ", client.IsDebug())
}
