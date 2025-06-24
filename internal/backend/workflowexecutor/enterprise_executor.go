//go:build enterprise
// +build enterprise

package workflowexecutor

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	MaxConcurrentWorkflows int                           `koanf:"max_concurrent_workflows"`
	WorkerID               string                        `koanf:"worker_id"`
	SessionWorkflow        temporalclient.WorkflowConfig `koanf:"session_workflow"`
}

var (
	Configs = configset.Set[Config]{
		Default: &Config{
			MaxConcurrentWorkflows: 1,
		},
		Test: &Config{
			WorkerID: "test-worker",
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

	inProgressWorkflowsCount int64
	cfg                      *Config

	executeLock sync.Mutex
}

func (e *executor) WorkflowQueue() string {
	return e.cfg.WorkerID + "-sessions-queue"
}

// Start implements WorkflowResourcesManager.
func (e *executor) Start(ctx context.Context) error {
	inProgress, err := e.svcs.DB.CountInProgressWorkflowExecutionRequests(ctx, e.cfg.WorkerID)
	if err != nil {
		return err
	}
	e.inProgressWorkflowsCount = inProgress
	e.l.Info("Starting workflow executor", zap.Int64("in_progress_workflows", e.inProgressWorkflowsCount))
	e.startPoller(ctx)
	return nil
}

func (e *executor) Stop(ctx context.Context) error {
	e.stopChannel <- struct{}{}
	return nil
}

func New(svcs Svcs, l *zap.Logger, cfg *Config) (*executor, error) {
	if cfg.WorkerID == "" {
		return nil, errors.New("worker_id is required")
	}

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
	}

	return e, nil
}

func (e *executor) availableSlots() int {
	return e.maxConcurrent - int(e.inProgressWorkflowsCount)
}

func (e *executor) Execute(ctx context.Context, sessionID sdktypes.SessionID, args any, memo map[string]string) error {
	return e.svcs.DB.CreateWorkflowExecutionRequest(ctx, db.WorkflowExecutionRequest{
		SessionID:  sessionID,
		WorkflowID: workflowID(sessionID),
		Args:       args,
		Memo:       memo,
	})
}

func (e *executor) startPoller(ctx context.Context) {
	timer := time.NewTimer(100 * time.Millisecond)

	go func() {
		for {
			select {
			case <-timer.C:
				e.runOnce(ctx)
				timer = time.NewTimer(100 * time.Millisecond) // Reset the timer for the next iteration
			case <-e.stopChannel:
				e.l.Info("Stopping workflow manager")
				timer.Stop()
				return
			case <-ctx.Done():
				return
			}

		}
	}()
}

func (e *executor) NotifyDone(ctx context.Context, id string) error {
	e.inProgressWorkflowsCount--
	if err := e.svcs.DB.UpdateRequestStatus(ctx, id, "done"); err != nil {
		e.l.Error("Failed to update workflow execution request status", zap.Error(err), zap.String("workflow_id", id))
	}
	e.l.Info(fmt.Sprintf("Active workflows: %d out of %d", e.inProgressWorkflowsCount, e.maxConcurrent))
	return nil
}

func (e *executor) runOnce(ctx context.Context) {
	e.executeLock.Lock()
	defer e.executeLock.Unlock()

	availableSlots := e.availableSlots()
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

		if err := e.svcs.DB.UpdateRequestStatus(ctx, job.WorkflowID, "in_progress"); err != nil {
			e.l.Error("Failed to delete workflow execution request", zap.Error(err), zap.String("session_id", job.SessionID.String()))
		}

		e.sampledLogger.Info(fmt.Sprintf("Active workflows: %d out of %d", e.inProgressWorkflowsCount, e.maxConcurrent))
	}
}

func (e *executor) executeAndIncrement(ctx context.Context, sessionID sdktypes.SessionID, workflowID string, args any, memo map[string]string) error {
	err := e.execute(ctx, sessionID, workflowID, args, memo)
	if err != nil {
		return err
	}
	e.inProgressWorkflowsCount++

	return nil
}
