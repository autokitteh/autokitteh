//go:build enterprise
// +build enterprise

package workflowresourcemanager

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/client"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type job struct {
	options client.StartWorkflowOptions
	name    string
	args    any
}

type manager struct {
	svcs  Svcs
	queue *q

	maxConcurrent int
	active        int
	l             *zap.Logger
	sampledLogger *zap.Logger

	stopChannel chan struct{}
}

// Start implements WorkflowResourcesManager.
func (e *manager) Start(ctx context.Context) error {
	e.startPoller(ctx)
	return nil
}

func (e *manager) Stop(ctx context.Context) error {
	e.stopChannel <- struct{}{}
	return nil
}

func New(svcs Svcs, l *zap.Logger) WorkflowResourcesManager {
	sampledLogger := l.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewSamplerWithOptions(core, time.Second, 1, 0)
	}))

	e := &manager{svcs: svcs,
		queue:         newQueue(svcs.DB),
		maxConcurrent: 1, active: 0,
		l:             l,
		sampledLogger: sampledLogger,
		stopChannel:   make(chan struct{}, 1),
	}

	return e
}

func (e *manager) Execute(ctx context.Context, options client.StartWorkflowOptions, name string, args any) error {
	e.queue.push(job{
		options: options,
		name:    name,
		args:    args,
	})
	return nil
}

func (e *manager) startPoller(ctx context.Context) {
	ticker := time.NewTicker(100 * time.Millisecond)
	go func() {
		for {
			select {
			case <-ticker.C:
				if e.queue.len() == 0 {
					e.sampledLogger.Debug("Queue is empty, waiting for jobs")
					continue
				}

				if e.active >= e.maxConcurrent {
					e.sampledLogger.Info("Max concurrent workflows reached, waiting for free slots", zap.Int("MaxConcurrent", e.maxConcurrent), zap.Int("QueueLength", e.queue.len()))
					continue
				}

				for _, job := range e.queue.popX(e.maxConcurrent - e.active) {
					id, err := e.execute(ctx, job.options, job.name, job.args)
					if err != nil {
						e.l.Error("Failed to execute workflow", zap.Error(err), zap.String("workflow_name", job.name), zap.Any("args", job.args))
						continue
					}
					e.l.Debug("Started workflow", zap.String("workflow_id", id), zap.String("workflow_name", job.name))
					e.active++
					e.l.Info(fmt.Sprintf("Active workflows: %d out of %d", e.active, e.maxConcurrent))
				}
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

func (e *manager) NotifyDone(ctx context.Context, id string) error {
	e.active--
	e.l.Info(fmt.Sprintf("Active workflows: %d out of %d", e.active, e.maxConcurrent))
	return nil
}

func (e *manager) execute(ctx context.Context, options client.StartWorkflowOptions, name string, args any) (string, error) {
	r, err := e.svcs.Temporal.TemporalClient().ExecuteWorkflow(
		ctx,
		options,
		name,
		args,
	)
	if err != nil {
		return "", err
	}
	return r.GetID(), nil
}
