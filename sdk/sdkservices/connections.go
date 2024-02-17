package sdkservices

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type ListConnectionsFilter struct {
	IntegrationID    sdktypes.IntegrationID
	IntegrationToken string
	ProjectID        sdktypes.ProjectID
}

type Connections interface {
	Create(ctx context.Context, conn sdktypes.Connection) (sdktypes.ConnectionID, error)
	Delete(ctx context.Context, id sdktypes.ConnectionID) error
	Update(ctx context.Context, conn sdktypes.Connection) error
	List(ctx context.Context, filter ListConnectionsFilter) ([]sdktypes.Connection, error)
	Get(ctx context.Context, id sdktypes.ConnectionID) (sdktypes.Connection, error)
}
