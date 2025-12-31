package main

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type SagaResult struct {
	OrderID      string
	Success      bool
	Steps        []StepResult
	Compensated  bool
	ErrorMessage string
}

func OrderSagaWorkflow(ctx workflow.Context, input SagaInput) (*SagaResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("OrderSagaWorkflow started", "orderID", input.OrderID)

	result := &SagaResult{
		OrderID: input.OrderID,
		Success: false,
		Steps:   []StepResult{},
	}

	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	var compensations []func(workflow.Context) error

	// Step 1
	var step1Result StepResult
	err := workflow.ExecuteActivity(ctx, Step1Activity, input).Get(ctx, &step1Result)
	if err != nil {
		logger.Error("Step1 failed", "error", err)
		result.ErrorMessage = fmt.Sprintf("Step1 failed: %v", err)
		return result, nil
	}
	result.Steps = append(result.Steps, step1Result)

	compensations = append(compensations, func(ctx workflow.Context) error {
		var compResult StepResult
		return workflow.ExecuteActivity(ctx, CompensateStep1Activity, input).Get(ctx, &compResult)
	})

	// Step 2
	var step2Result StepResult
	err = workflow.ExecuteActivity(ctx, Step2Activity, input).Get(ctx, &step2Result)
	if err != nil {
		logger.Error("Step2 failed, running compensations", "error", err)
		result.ErrorMessage = fmt.Sprintf("Step2 failed: %v", err)
		runCompensations(ctx, logger, compensations)
		result.Compensated = true
		return result, nil
	}
	result.Steps = append(result.Steps, step2Result)

	compensations = append(compensations, func(ctx workflow.Context) error {
		var compResult StepResult
		return workflow.ExecuteActivity(ctx, CompensateStep2Activity, input).Get(ctx, &compResult)
	})

	// Step 3
	var step3Result StepResult
	err = workflow.ExecuteActivity(ctx, Step3Activity, input).Get(ctx, &step3Result)
	if err != nil {
		logger.Error("Step3 failed, running compensations", "error", err)
		result.ErrorMessage = fmt.Sprintf("Step3 failed: %v", err)
		runCompensations(ctx, logger, compensations)
		result.Compensated = true
		return result, nil
	}
	result.Steps = append(result.Steps, step3Result)

	result.Success = true
	logger.Info("OrderSagaWorkflow completed successfully", "orderID", input.OrderID)
	return result, nil
}

func runCompensations(ctx workflow.Context, logger log.Logger, compensations []func(workflow.Context) error) {
	logger.Info("Running compensations", "count", len(compensations))
	for i := len(compensations) - 1; i >= 0; i-- {
		err := compensations[i](ctx)
		if err != nil {
			logger.Error("Compensation failed", "index", i, "error", err)
		}
	}
}
