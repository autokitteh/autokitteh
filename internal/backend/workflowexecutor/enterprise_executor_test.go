//go:build enterprise
// +build enterprise

package workflowexecutor

import (
	"context"
	"testing"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.temporal.io/sdk/client"
	"go.uber.org/zap"
	"gotest.tools/v3/assert"
)

func TestAvailableSlots(t *testing.T) {
	e := executor{maxConcurrent: 10, inProgressWorkflowsCount: 10}

	e.inProgressWorkflowsCount = int64(e.maxConcurrent)

	assert.Equal(t, e.availableSlots(), 0, "Expected no available slots when all slots are in use")

	e.inProgressWorkflowsCount = int64(e.maxConcurrent - 1)
	assert.Equal(t, e.availableSlots(), 1, "Expected one available slot when one slot is free")
}

func TestRunOnceNoAvailableSlots(t *testing.T) {
	mdb := &mockDB{}
	e := getExecutor(
		mdb,
		nil,
		&Config{
			MaxConcurrentWorkflows: 10,
			WorkerID:               "test-worker",
		},
	)

	e.inProgressWorkflowsCount = int64(e.maxConcurrent)

	e.runOnce(t.Context())

	assert.Equal(t, mdb.getRequestCount, 0, "Expected no requests to be made when no slots are available")
}

func TestRunOnceOneJob(t *testing.T) {
	sid := sdktypes.NewSessionID()

	mdb := &mockDB{
		dbResult: func() ([]db.WorkflowExecutionRequest, error) {
			return []db.WorkflowExecutionRequest{
				{
					SessionID:  sid,
					WorkflowID: "test-workflow",
					Args:       map[string]interface{}{"key": "value"},
					Memo:       map[string]string{"memoKey": "memoValue"},
				},
			}, nil
		},
	}

	mockTemporal := mockTemporalClient{}

	e := getExecutor(
		mdb,
		&mockTemporal,
		&Config{
			MaxConcurrentWorkflows: 1,
			WorkerID:               "test-worker",
			SessionWorkflow:        temporalclient.WorkflowConfig{},
		},
	)

	e.runOnce(t.Context())

	// Verify DB
	assert.Equal(t, mdb.getRequestCount, 1, "Expected one request to be made")
	assert.Equal(t, mdb.updateRequestStatusCallCount, 1, "Expected request status to be updated once")

	// Verify Temporal
	assert.Equal(t, mockTemporal.executeWorkflowCallCount, 1, "Expected workflow to be executed once")
	assert.Equal(t, mockTemporal.executeWorkflowName, e.WorkflowSessionName(), "Expected workflow name to match")

	// Verify executor
	assert.Equal(t, e.inProgressWorkflowsCount, int64(1), "Expected in-progress workflows count to be incremented")

	// Test we don't take more jobs if we have no slots available
	e.runOnce(t.Context())
	assert.Equal(t, mdb.getRequestCount, 1, "Expected one request to be made")

	//
	e.maxConcurrent = e.maxConcurrent + 1
	e.runOnce(t.Context())
	assert.Equal(t, mdb.getRequestCount, 2, "Expected one request to be made")

	assert.Equal(t, e.inProgressWorkflowsCount, int64(2), "Expected in-progress workflows count to be incremented")

	err := e.NotifyDone(t.Context(), "test-workflow")
	assert.NilError(t, err, "Expected NotifyDone to succeed")
	assert.Equal(t, e.inProgressWorkflowsCount, int64(1), "Expected in-progress workflows count to be decremented")
}

// Utilities and mocks for testing
type executeResult struct {
	client.WorkflowRun
}

func (r executeResult) GetID() string {
	return "mock-workflow-id"
}

type mockTemporalClient struct {
	client.Client
	executeWorkflowCallCount int
	executeWorkflowName      string
	executeArgs              interface{}
}

func (m *mockTemporalClient) ExecuteWorkflow(ctx context.Context, options client.StartWorkflowOptions, workflowType interface{}, args ...interface{}) (client.WorkflowRun, error) {
	m.executeWorkflowCallCount++
	m.executeWorkflowName = workflowType.(string)
	m.executeArgs = args
	return executeResult{}, nil // Mock implementation, should return a valid WorkflowRun
}

type mockTemporalSvc struct {
	temporalclient.Client
	client client.Client
}

func (m mockTemporalSvc) TemporalClient() client.Client {
	return m.client
}

type mockDB struct {
	db.DB
	getRequestCount int
	getRequestArgs  struct {
		workerID string
		slots    int
	}

	updateRequestStatusCallCount int

	dbResult func() ([]db.WorkflowExecutionRequest, error)
}

func (m *mockDB) GetWorkflowExecutionRequests(ctx context.Context, workerID string, maxRequests int) ([]db.WorkflowExecutionRequest, error) {
	m.getRequestCount++
	m.getRequestArgs.workerID = workerID
	m.getRequestArgs.slots = maxRequests

	return m.dbResult()
}
func (m *mockDB) UpdateRequestStatus(ctx context.Context, workflowID string, status string) error {
	m.updateRequestStatusCallCount++
	// Mock implementation, just return nil to simulate success
	return nil
}

func getExecutor(
	dbMock db.DB,
	temporal *mockTemporalClient,
	cfg *Config,
) *executor {

	svcs := Svcs{
		DB:       dbMock,                            // Mock or real DB instance
		Temporal: mockTemporalSvc{client: temporal}, // Mock or real Temporal client
	}

	e, _ := New(svcs,
		zap.NewNop(),
		cfg,
	)

	return e
}
