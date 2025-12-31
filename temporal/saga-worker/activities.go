package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"
)

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

func Step1Activity(ctx context.Context, input SagaInput) (*StepResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Step1Activity started", "orderID", input.OrderID)
	time.Sleep(500 * time.Millisecond)

	if input.ShouldFail == "step1" {
		return nil, errors.New("Step1 failed: simulated error")
	}

	return &StepResult{
		StepName:  "Step1",
		Success:   true,
		Message:   fmt.Sprintf("Order %s created", input.OrderID),
		Timestamp: time.Now(),
	}, nil
}

func Step2Activity(ctx context.Context, input SagaInput) (*StepResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Step2Activity started", "orderID", input.OrderID)
	time.Sleep(500 * time.Millisecond)

	if input.ShouldFail == "step2" {
		return nil, errors.New("Step2 failed: simulated error")
	}

	return &StepResult{
		StepName:  "Step2",
		Success:   true,
		Message:   fmt.Sprintf("Payment for %s processed", input.OrderID),
		Timestamp: time.Now(),
	}, nil
}

func Step3Activity(ctx context.Context, input SagaInput) (*StepResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Step3Activity started", "orderID", input.OrderID)
	time.Sleep(500 * time.Millisecond)

	if input.ShouldFail == "step3" {
		return nil, errors.New("Step3 failed: simulated error")
	}

	return &StepResult{
		StepName:  "Step3",
		Success:   true,
		Message:   fmt.Sprintf("Inventory for %s reserved", input.OrderID),
		Timestamp: time.Now(),
	}, nil
}

func CompensateStep1Activity(ctx context.Context, input SagaInput) (*StepResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("CompensateStep1 - Cancelling order", "orderID", input.OrderID)
	time.Sleep(300 * time.Millisecond)

	return &StepResult{
		StepName:  "CompensateStep1",
		Success:   true,
		Message:   fmt.Sprintf("Order %s cancelled", input.OrderID),
		Timestamp: time.Now(),
	}, nil
}

func CompensateStep2Activity(ctx context.Context, input SagaInput) (*StepResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("CompensateStep2 - Refunding payment", "orderID", input.OrderID)
	time.Sleep(300 * time.Millisecond)

	return &StepResult{
		StepName:  "CompensateStep2",
		Success:   true,
		Message:   fmt.Sprintf("Payment for %s refunded", input.OrderID),
		Timestamp: time.Now(),
	}, nil
}

func CompensateStep3Activity(ctx context.Context, input SagaInput) (*StepResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("CompensateStep3 - Releasing inventory", "orderID", input.OrderID)
	time.Sleep(300 * time.Millisecond)

	return &StepResult{
		StepName:  "CompensateStep3",
		Success:   true,
		Message:   fmt.Sprintf("Inventory for %s released", input.OrderID),
		Timestamp: time.Now(),
	}, nil
}
