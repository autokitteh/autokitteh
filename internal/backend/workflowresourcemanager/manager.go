//go:build !enterprise
// +build !enterprise

package workflowresourcemanager

import (
	"context"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.temporal.io/sdk/client"
	"go.uber.org/zap"
)

type Config struct {
}

var (
	Configs = configset.Set[Config]{}
)

type manager struct {
	svcs Svcs
	l    *zap.Logger
}

func New(svcs Svcs, l *zap.Logger) WorkflowResourcesManager {
	return &manager{svcs: svcs, l: l}
}

func (e *manager) Execute(ctx context.Context, options client.StartWorkflowOptions, name string, args any) error {
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

func (e *manager) NotifyDone(ctx context.Context, id string) error {
	return nil
}

func (e *manager) Start(ctx context.Context) error {

	return nil
}

func (e *manager) Stop(ctx context.Context) error {

	return nil
}
