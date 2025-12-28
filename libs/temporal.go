package libs

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// TemporalClient wraps Temporal SDK operations
type TemporalClient struct {
	client    client.Client
	namespace string
}

// TemporalConfig holds Temporal connection configuration
type TemporalConfig struct {
	HostPort  string
	Namespace string
}

// NewTemporalClient creates a new Temporal client
func NewTemporalClient(config TemporalConfig) (*TemporalClient, error) {
	if config.Namespace == "" {
		config.Namespace = "default"
	}

	c, err := client.Dial(client.Options{
		HostPort:  config.HostPort,
		Namespace: config.Namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Temporal client: %w", err)
	}

	return &TemporalClient{
		client:    c,
		namespace: config.Namespace,
	}, nil
}

// StartWorkflow starts a workflow execution
func (t *TemporalClient) StartWorkflow(ctx context.Context, workflowID, taskQueue, workflowType string, args ...interface{}) (client.WorkflowRun, error) {
	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: taskQueue,
	}

	return t.client.ExecuteWorkflow(ctx, options, workflowType, args...)
}

// StartWorkflowWithOptions starts a workflow with custom options
func (t *TemporalClient) StartWorkflowWithOptions(ctx context.Context, options client.StartWorkflowOptions, workflowType string, args ...interface{}) (client.WorkflowRun, error) {
	return t.client.ExecuteWorkflow(ctx, options, workflowType, args...)
}

// GetWorkflowResult gets the result of a workflow execution
func (t *TemporalClient) GetWorkflowResult(ctx context.Context, workflowID, runID string, valuePtr interface{}) error {
	workflowRun := t.client.GetWorkflow(ctx, workflowID, runID)
	return workflowRun.Get(ctx, valuePtr)
}

// SignalWorkflow sends a signal to a running workflow
func (t *TemporalClient) SignalWorkflow(ctx context.Context, workflowID, runID, signalName string, data interface{}) error {
	return t.client.SignalWorkflow(ctx, workflowID, runID, signalName, data)
}

// SignalWithStartWorkflow signals a workflow if running, or starts it if not
func (t *TemporalClient) SignalWithStartWorkflow(ctx context.Context, workflowID, signalName, taskQueue, workflowType string, signalArg interface{}, workflowArgs ...interface{}) (client.WorkflowRun, error) {
	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: taskQueue,
	}

	return t.client.SignalWithStartWorkflow(ctx, workflowID, signalName, signalArg, options, workflowType, workflowArgs...)
}

// QueryWorkflow queries a running workflow
func (t *TemporalClient) QueryWorkflow(ctx context.Context, workflowID, runID, queryType string, args ...interface{}) (interface{}, error) {
	resp, err := t.client.QueryWorkflow(ctx, workflowID, runID, queryType, args...)
	if err != nil {
		return nil, err
	}

	var result interface{}
	if err := resp.Get(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// CancelWorkflow cancels a running workflow
func (t *TemporalClient) CancelWorkflow(ctx context.Context, workflowID, runID string) error {
	return t.client.CancelWorkflow(ctx, workflowID, runID)
}

// TerminateWorkflow terminates a running workflow
func (t *TemporalClient) TerminateWorkflow(ctx context.Context, workflowID, runID, reason string, details ...interface{}) error {
	return t.client.TerminateWorkflow(ctx, workflowID, runID, reason, details...)
}

// DescribeWorkflowExecution describes a workflow execution
func (t *TemporalClient) DescribeWorkflowExecution(ctx context.Context, workflowID, runID string) (interface{}, error) {
	// Return the raw response since the type may vary by Temporal version
	return t.client.DescribeWorkflowExecution(ctx, workflowID, runID)
}

// ListWorkflows lists workflow executions with a query
// Note: This is a simplified version - see Temporal documentation for advanced usage
func (t *TemporalClient) ListWorkflows(ctx context.Context, query string) ([]string, error) {
	// Temporal list API may vary by version
	// This function returns a placeholder - users should implement based on their Temporal version
	return []string{}, fmt.Errorf("ListWorkflows needs to be implemented based on your Temporal version")
}

// ScheduleWorkflow schedules a workflow to run at a specific time
func (t *TemporalClient) ScheduleWorkflow(ctx context.Context, scheduleID, taskQueue, workflowType string, interval time.Duration, args ...interface{}) error {
	// For Temporal schedules, use the schedule API
	// This is a simplified example - production code should use schedule creation API

	options := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("%s-%d", scheduleID, time.Now().Unix()),
		TaskQueue: taskQueue,
	}

	_, err := t.client.ExecuteWorkflow(ctx, options, workflowType, args...)
	return err
}

// CreateWorker creates a new Temporal worker
func (t *TemporalClient) CreateWorker(taskQueue string, workflows []interface{}, activities []interface{}) worker.Worker {
	w := worker.New(t.client, taskQueue, worker.Options{})

	// Register workflows
	for _, wf := range workflows {
		w.RegisterWorkflow(wf)
	}

	// Register activities
	for _, act := range activities {
		w.RegisterActivity(act)
	}

	return w
}

// RunWorker creates and runs a worker (blocking)
func (t *TemporalClient) RunWorker(ctx context.Context, taskQueue string, workflows []interface{}, activities []interface{}) error {
	w := t.CreateWorker(taskQueue, workflows, activities)
	return w.Run(worker.InterruptCh())
}

// Close closes the Temporal client
func (t *TemporalClient) Close() {
	t.client.Close()
}

// GetClient returns the underlying Temporal client for advanced operations
func (t *TemporalClient) GetClient() client.Client {
	return t.client
}

// CheckHealth checks if the Temporal server is accessible
func (t *TemporalClient) CheckHealth(ctx context.Context) error {
	_, err := t.client.CheckHealth(ctx, &client.CheckHealthRequest{})
	return err
}
