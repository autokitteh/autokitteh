//go:build enterprise
// +build enterprise

package workflowexecutor

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	v11 "go.temporal.io/api/enums/v1"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Config struct {
	MaxConcurrentWorkflows int                           `koanf:"max_concurrent_workflows"`
	WorkerID               string                        `koanf:"worker_id"`
	SessionWorkflow        temporalclient.WorkflowConfig `koanf:"session_workflow"`
	EnablePoller           bool                          `koanf:"enable_poller"`
	PollerIntervalMS       time.Duration                 `koanf:"poller_interval_ms"`
	ReconcileLoopInterval  time.Duration                 `koanf:"reconcile_loop_interval"`
}

var (
	Configs = configset.Set[Config]{
		Default: &Config{
			MaxConcurrentWorkflows: 1,
			EnablePoller:           false,
			PollerIntervalMS:       100 * time.Millisecond,
			ReconcileLoopInterval:  1 * time.Hour,
		},
		Test: &Config{
			WorkerID:     "test-worker",
			EnablePoller: true,
		},
		Dev: &Config{
			MaxConcurrentWorkflows: 10,
			EnablePoller:           true,
			WorkerID:               "dev-worker",
		},
	}
)

var _ WorkflowExecutor = (*executor)(nil)

type executor struct {
	svcs Svcs

	maxConcurrent int
	l             *zap.Logger
	sampledLogger *zap.Logger

	stopChannel chan struct{}

	inProgressWorkflowsCount atomic.Int64
	cfg                      *Config

	executeLock sync.Mutex
	metrics     *metrics
}

func (e *executor) WorkflowQueue() string {
	return e.cfg.WorkerID + "-sessions-queue"
}

// Start implements WorkflowResourcesManager.
func (e *executor) Start(ctx context.Context) error {
	if !e.cfg.EnablePoller {
		e.l.Info("Workflow executor polling is disabled, skipping start")
		return nil
	}

	if e.cfg.WorkerID == "" {
		return errors.New("worker_id is required")
	}

	// we do this synchronous just to ensure everything is aligned on start
	e.reconcile(ctx)
	e.startReconcileLoop(ctx)

	inProgress, err := e.svcs.DB.CountInProgressWorkflowExecutionRequests(ctx, e.cfg.WorkerID)
	if err != nil {
		return err
	}
	e.inProgressWorkflowsCount.Store(inProgress)
	e.l.Info(fmt.Sprintf("Starting workflow executor. in_progress_workflows: %d", inProgress))
	e.startPoller(ctx)
	return nil
}

func (e *executor) startReconcileLoop(ctx context.Context) {
	go func() {
		e.l.Info("Starting workflow executor reconcile loop")
		for {
			select {
			case <-time.After(e.cfg.ReconcileLoopInterval):
				e.reconcile(ctx)
			case <-e.stopChannel:
				e.l.Info("Stopping workflow executor reconcile loop")
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (e *executor) reconcile(ctx context.Context) error {
	workflowIds, err := e.svcs.DB.GetInProgressWorkflowIDs(ctx, e.cfg.WorkerID)
	if err != nil {
		return err
	}

	if len(workflowIds) == 0 {
		return nil
	}

	e.l.Info("Reconciling in-progress workflows", zap.Int("count", len(workflowIds)))

	for _, workflowId := range workflowIds {
		resp, err := e.svcs.Temporal.TemporalClient().DescribeWorkflowExecution(ctx, workflowId, "")
		if err != nil {
			e.l.Error("Failed to describe workflow execution", zap.Error(err), zap.String("workflow_id", workflowId))
			continue
		}

		if resp.WorkflowExecutionInfo.Status != v11.WORKFLOW_EXECUTION_STATUS_RUNNING {
			e.l.Info("Workflow is no longer running, releasing", zap.String("workflow_id", workflowId))
			if err := e.NotifyDone(ctx, workflowId); err != nil {
				e.l.Error("Failed to notify workflow done", zap.Error(err), zap.String("workflow_id", workflowId))
			} else {
				e.l.Info("Workflow execution request status updated to done", zap.String("workflow_id", workflowId))
			}
		}
	}
	return nil
}

func (e *executor) Stop(ctx context.Context) error {
	e.stopChannel <- struct{}{}
	return nil
}

func New(svcs Svcs, l *zap.Logger, cfg *Config) (*executor, error) {
	sampledLogger := l.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewSamplerWithOptions(core, time.Second, 1, 0)
	}))

	e := &executor{svcs: svcs,
		maxConcurrent: cfg.MaxConcurrentWorkflows,
		l:             l,
		sampledLogger: sampledLogger,
		stopChannel:   make(chan struct{}, 1),
		cfg:           cfg,
		executeLock:   sync.Mutex{},
		metrics:       newMetrics(cfg.WorkerID),
	}

	return e, nil
}

func (e *executor) availableSlots(ctx context.Context) int {
	active := int(e.inProgressWorkflowsCount.Load())
	max := e.maxConcurrent
	load := (float64(active) * 100.0) / float64(max)
	e.metrics.SetWorkerLoad(ctx, load)
	return max - active
}

func (e *executor) Execute(ctx context.Context, sessionID sdktypes.SessionID, args any, memo map[string]string) error {
	if err := e.svcs.DB.CreateWorkflowExecutionRequest(ctx, db.WorkflowExecutionRequest{
		SessionID:  sessionID,
		WorkflowID: workflowID(sessionID),
		Args:       args,
		Memo:       memo,
	}); err != nil {
		return err
	}
	e.metrics.IncrementQueuedWorkflows(ctx)
	return nil
}

func (e *executor) startPoller(ctx context.Context) {
	go func() {
		for {
			select {
			case <-time.After(e.cfg.PollerIntervalMS):
				e.runOnce(ctx)
			case <-e.stopChannel:
				e.l.Info("Stopping workflow manager")
				return
			case <-ctx.Done():
				return
			}

		}
	}()
}

func (e *executor) NotifyDone(ctx context.Context, id string) error {
	didUpdate, err := e.svcs.DB.UpdateWorkflowExecutionRequestStatus(ctx, id, "done")
	if err != nil {
		e.l.Error("Failed to update workflow execution request status", zap.Error(err), zap.String("workflow_id", id))
		return err
	}

	if didUpdate {
		e.inProgressWorkflowsCount.Add(-1)
		e.metrics.DecrementActiveWorkflows(ctx)
		e.l.Info(fmt.Sprintf("Workflow Done %s. Active workflows: %d out of %d", id, e.inProgressWorkflowsCount.Load(), e.maxConcurrent), zap.String("session_id", id))
	}

	return nil
}

func (e *executor) runOnce(ctx context.Context) {
	e.executeLock.Lock()
	defer e.executeLock.Unlock()

	availableSlots := e.availableSlots(ctx)
	if availableSlots <= 0 {
		e.sampledLogger.Info("Max concurrent workflows reached, waiting for free slots", zap.Int("MaxConcurrent", e.maxConcurrent))
		return
	}

	requests, err := e.svcs.DB.GetWorkflowExecutionRequests(ctx, e.cfg.WorkerID, availableSlots)
	if err != nil {
		e.l.Error("Failed to get workflow execution requests", zap.Error(err))
		return
	}

	for _, job := range requests {
		err := e.executeAndIncrement(ctx, job.SessionID, job.WorkflowID, job.Args, job.Memo)
		if err != nil {
			e.l.Error("Failed to execute workflow", zap.Error(err), zap.String("workflow_name", e.WorkflowSessionName()))
			continue
		}

		if job.RetryCount == 1 {
			// we don't want to dequeue it twice if we retry it
			e.metrics.DecrementQueuedWorkflows(ctx)
		} else {
			e.metrics.IncrementRetriedWorkflowsCounter(ctx)
		}

		e.sampledLogger.Info(fmt.Sprintf("Workflow Started. Active workflows: %d out of %d", e.inProgressWorkflowsCount.Load(), e.maxConcurrent))
	}
}

func (e *executor) executeAndIncrement(ctx context.Context, sessionID sdktypes.SessionID, workflowID string, args any, memo map[string]string) error {
	err := e.execute(ctx, sessionID, workflowID, args, memo)
	if err != nil {
		return err
	}
	e.inProgressWorkflowsCount.Add(1)

	e.metrics.IncrementActiveWorkflows(ctx)

	return nil
}
