package workflowexecutor

import (
	"context"

	"go.temporal.io/sdk/client"
	"go.uber.org/fx"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
)

type Svcs struct {
	fx.In
	Temporal temporalclient.Client
	DB       db.DB
}

type WorkflowExecutor interface {
	Execute(ctx context.Context, options client.StartWorkflowOptions, name string, args any) error
	NotifyDone(ctx context.Context, id string) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
