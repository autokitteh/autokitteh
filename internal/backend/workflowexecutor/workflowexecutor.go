package workflowexecutor

import (
	"context"

	"go.uber.org/fx"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Svcs struct {
	fx.In
	Temporal temporalclient.Client
	DB       db.DB
}

// TODO: need to handle child workflows
type WorkflowExecutor interface {
	Execute(ctx context.Context, sessionID sdktypes.SessionID, data any, memo map[string]string) error
	NotifyDone(ctx context.Context, id string) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error

	WorkflowQueue() string
	WorkflowSessionName() string
}
