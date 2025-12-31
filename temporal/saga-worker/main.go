package main

import (
	"log"
	"os"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

const TaskQueue = "saga-task-queue"

func main() {
	temporalAddr := os.Getenv("TEMPORAL_ADDRESS")
	if temporalAddr == "" {
		temporalAddr = "localhost:7233"
	}

	log.Printf("Connecting to Temporal at %s", temporalAddr)

	c, err := client.Dial(client.Options{
		HostPort: temporalAddr,
	})
	if err != nil {
		log.Fatalf("Failed to create Temporal client: %v", err)
	}
	defer c.Close()

	log.Println("Successfully connected to Temporal")

	w := worker.New(c, TaskQueue, worker.Options{})

	w.RegisterWorkflow(OrderSagaWorkflow)
	w.RegisterActivity(Step1Activity)
	w.RegisterActivity(Step2Activity)
	w.RegisterActivity(Step3Activity)
	w.RegisterActivity(CompensateStep1Activity)
	w.RegisterActivity(CompensateStep2Activity)
	w.RegisterActivity(CompensateStep3Activity)

	log.Printf("Starting worker on task queue: %s", TaskQueue)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalf("Failed to start worker: %v", err)
	}
}
