//go:build !enterprise
// +build !enterprise

package workflowresourcemanager

import (
	"context"

	"go.temporal.io/sdk/client"
	"go.uber.org/zap"
)

type executer struct {
	svcs Svcs
	l    *zap.Logger
}

func New(svcs Svcs, l *zap.Logger) WorkflowResourcesManager {
	return &executer{svcs: svcs, l: l}
}

func (e *executer) Execute(ctx context.Context, options client.StartWorkflowOptions, name string, args any) error {
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

func (e *executer) NotifyDone(ctx context.Context, id string) error {
	return nil
}

func (e *executer) Start(ctx context.Context) error {

	return nil
}

func (e *executer) Stop(ctx context.Context) error {

	return nil
}
