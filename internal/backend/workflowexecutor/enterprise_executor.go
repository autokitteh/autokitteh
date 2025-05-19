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
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.temporal.io/sdk/client"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	MaxConcurrentWorkflows int    `koanf:"max_concurrent_workflows"`
	WorkerID               string `koanf:"worker_id"`
}

var (
	Configs = configset.Set[Config]{
		Default: &Config{
			MaxConcurrentWorkflows: 1,
		},
	}
)

type job struct {
	options client.StartWorkflowOptions
	name    string
	args    any
}

type executor struct {
	svcs  Svcs
	queue *q

	maxConcurrent int
	l             *zap.Logger
	sampledLogger *zap.Logger

	stopChannel chan struct{}

	workerInfo scheme.WorkerInfo
	cfg        *Config

	executeLock *sync.Mutex
}

// Start implements WorkflowResourcesManager.
func (e *executor) Start(ctx context.Context) error {
	workerInfo, err := e.svcs.DB.GetWorkerInfo(ctx, e.cfg.WorkerID)
	if err != nil {
		return err
	}
	e.workerInfo = workerInfo
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
		queue:         newQueue(svcs.DB),
		maxConcurrent: cfg.MaxConcurrentWorkflows,
		l:             l,
		sampledLogger: sampledLogger,
		stopChannel:   make(chan struct{}, 1),
		cfg:           cfg,
		executeLock:   &sync.Mutex{},
	}

	return e, nil
}

func (e *executor) availableSlots() int {
	return e.maxConcurrent - e.workerInfo.ActiveWorkflows
}

func (e *executor) Execute(ctx context.Context, options client.StartWorkflowOptions, name string, args any) error {
	// by pass queue if we have free slots
	if ok := e.executeLock.TryLock(); ok {

		defer e.executeLock.Unlock()
		if e.availableSlots() > 0 {
			_, err := e.execute(ctx, options, name, args)
			return err
		}
	}

	e.queue.push(job{
		options: options,
		name:    name,
		args:    args,
	})

	return nil
}

func (e *executor) startPoller(ctx context.Context) {
	ticker := time.NewTicker(100 * time.Millisecond)
	go func() {
		for {
			select {
			case <-ticker.C:
				e.runOnce(ctx)
			case <-e.stopChannel:
				e.l.Info("Stopping workflow manager")
				ticker.Stop()
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (e *executor) NotifyDone(ctx context.Context, id string) error {
	var err error
	e.workerInfo.ActiveWorkflows, err = e.svcs.DB.DecActiveWorkflows(ctx, e.cfg.WorkerID)
	if err != nil {
		e.l.Error("Failed to update active workflows", zap.Error(err), zap.String("workflow_id", id))
		e.workerInfo.ActiveWorkflows--
	}
	e.l.Info(fmt.Sprintf("Active workflows: %d out of %d", e.workerInfo.ActiveWorkflows, e.maxConcurrent))
	return nil
}

func (e *executor) runOnce(ctx context.Context) {
	e.executeLock.Lock()
	defer e.executeLock.Unlock()

	if e.queue.len() == 0 {
		e.sampledLogger.Debug("Queue is empty, waiting for jobs")
		return
	}

	availableSlots := e.availableSlots()
	if availableSlots <= 0 {
		e.sampledLogger.Info("Max concurrent workflows reached, waiting for free slots", zap.Int("MaxConcurrent", e.maxConcurrent), zap.Int("QueueLength", e.queue.len()))
		return
	}

	for _, job := range e.queue.popX(availableSlots) {
		id, err := e.execute(ctx, job.options, job.name, job.args)
		if err != nil {
			e.l.Error("Failed to execute workflow", zap.Error(err), zap.String("workflow_name", job.name), zap.Any("args", job.args))
			continue
		}

		e.l.Debug("Started workflow", zap.String("workflow_id", id), zap.String("workflow_name", job.name))
		e.l.Info(fmt.Sprintf("Active workflows: %d out of %d", e.workerInfo.ActiveWorkflows, e.maxConcurrent))
	}
}

func (e *executor) execute(ctx context.Context, options client.StartWorkflowOptions, name string, args any) (string, error) {
	r, err := e.svcs.Temporal.TemporalClient().ExecuteWorkflow(
		ctx,
		options,
		name,
		args,
	)
	if err != nil {
		return "", err
	}

	e.workerInfo.ActiveWorkflows, err = e.svcs.DB.IncActiveWorkflows(ctx, e.cfg.WorkerID)
	if err != nil {
		e.l.Error("Failed to update active workflows", zap.Error(err))
		// This is so at least the in memory represination would be ok
		// we might succeed updating later
		e.workerInfo.ActiveWorkflows++
	}
	return r.GetID(), nil
}
