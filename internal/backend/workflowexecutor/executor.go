//go:build !enterprise
// +build !enterprise

package workflowexecutor

import (
	"context"

	"go.temporal.io/sdk/client"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
)

type Config struct {
}

var (
	Configs = configset.Set[Config]{
		Default: &Config{},
	}
)

type executor struct {
	svcs Svcs
	l    *zap.Logger
}

func New(svcs Svcs, l *zap.Logger) *executor {
	return &executor{svcs: svcs, l: l}
}

func (e *executor) Execute(ctx context.Context, options client.StartWorkflowOptions, name string, args any) error {
	r, err := e.svcs.Temporal.TemporalClient().ExecuteWorkflow(
		ctx,
		options,
		name,
		args,
	)
	if err != nil {
		return err
	}
	e.l.Info("Started workflow", zap.String("workflow_id", r.GetID()), zap.String("workflow_name", name))
	return nil
}

func (e *executor) NotifyDone(ctx context.Context, id string) error {
	return nil
}

func (e *executor) Start(ctx context.Context) error {

	return nil
}

func (e *executor) Stop(ctx context.Context) error {

	return nil
}
