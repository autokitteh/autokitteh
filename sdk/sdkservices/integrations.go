package sdkservices

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdkexecutor"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// Integrations is implemented by the autokitteh core, to manage integrations
// of external services within a running autokitteh server.
type Integrations interface {
	// Get returns the instance of an integration which has already been registered
	// in the autokitteh server, and is available for usage by runtime connections.
	GetByID(ctx context.Context, id sdktypes.IntegrationID) (Integration, error)
	GetByName(ctx context.Context, name sdktypes.Symbol) (Integration, error)

	// List returns an enumeration - with optional filtering - of all
	// the integrations which have been registered in the autokitteh
	// server, and are available for usage by runtime connections.
	// TODO: Add an optional tag-search-term filter argument.
	List(ctx context.Context, nameSubstring string) ([]sdktypes.Integration, error)
}

// Integration is implemented for each external service, to let the autokitteh
// server interact with it.
type Integration interface {
	// Get returns the configuration details of this integration and the
	// external service that it wraps.
	Get() sdktypes.Integration

	Configure(ctx context.Context, cid sdktypes.ConnectionID) (map[string]sdktypes.Value, map[string]string, error)

	TestConnection(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error)

	GetConnectionStatus(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error)

	GetConnectionConfig(ctx context.Context, cid sdktypes.ConnectionID) (map[string]string, error)

	sdkexecutor.Caller
}
