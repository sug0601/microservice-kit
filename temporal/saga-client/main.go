package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"go.temporal.io/sdk/client"
)

const TaskQueue = "saga-task-queue"

type SagaInput struct {
	OrderID    string
	ShouldFail string
}

type StepResult struct {
	StepName  string
	Success   bool
	Message   string
	Timestamp time.Time
}

type SagaResult struct {
	OrderID      string
	Success      bool
	Steps        []StepResult
	Compensated  bool
	ErrorMessage string
}

func main() {
	failAt := flag.String("fail", "", "Force failure at step (step1, step2, step3)")
	orderID := flag.String("order", "", "Order ID (auto-generated if not specified)")
	flag.Parse()

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

	if *orderID == "" {
		*orderID = fmt.Sprintf("order-%d", time.Now().UnixNano())
	}

	input := SagaInput{
		OrderID:    *orderID,
		ShouldFail: *failAt,
	}

	log.Printf("Starting OrderSaga workflow")
	log.Printf("  Order ID: %s", input.OrderID)
	if input.ShouldFail != "" {
		log.Printf("  Simulating failure at: %s", input.ShouldFail)
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("saga-%s", input.OrderID),
		TaskQueue: TaskQueue,
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, "OrderSagaWorkflow", input)
	if err != nil {
		log.Fatalf("Failed to start workflow: %v", err)
	}

	log.Printf("Workflow started: WorkflowID=%s, RunID=%s", we.GetID(), we.GetRunID())

	var result SagaResult
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalf("Workflow failed: %v", err)
	}

	fmt.Println("\n========================================")
	fmt.Println("           SAGA RESULT")
	fmt.Println("========================================")
	fmt.Printf("Order ID:    %s\n", result.OrderID)
	fmt.Printf("Success:     %v\n", result.Success)
	fmt.Printf("Compensated: %v\n", result.Compensated)
	if result.ErrorMessage != "" {
		fmt.Printf("Error:       %s\n", result.ErrorMessage)
	}

	fmt.Println("\nSteps executed:")
	for i, step := range result.Steps {
		status := "✓"
		if !step.Success {
			status = "✗"
		}
		fmt.Printf("  %d. [%s] %s: %s\n", i+1, status, step.StepName, step.Message)
	}
	fmt.Println("========================================")

	jsonResult, _ := json.MarshalIndent(result, "", "  ")
	log.Printf("Full result (JSON):\n%s", string(jsonResult))
}
