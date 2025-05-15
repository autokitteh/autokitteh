package workflowresourcemanager

import (
	"context"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.temporal.io/sdk/client"
	"go.uber.org/fx"
)

type Svcs struct {
	fx.In
	Temporal temporalclient.Client
	DB       db.DB
}

type WorkflowResourcesManager interface {
	Execute(ctx context.Context, options client.StartWorkflowOptions, name string, args any) error
	NotifyDone(ctx context.Context, id string) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
