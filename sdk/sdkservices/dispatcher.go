package sdkservices

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type DispatchOptions struct {
	// If set, dispatch only to this specific deployment.
	DeploymentID sdktypes.DeploymentID

	// If true, the call will block until the dispatch is done.
	Wait bool
}

type DispatchResponse struct {
	EventID sdktypes.EventID

	// Returned only if Wait was true.
	SessionIDs []sdktypes.SessionID
}
type Dispatcher interface {
	Dispatch(ctx context.Context, event sdktypes.Event, opts *DispatchOptions) (*DispatchResponse, error)
	Redispatch(ctx context.Context, eventID sdktypes.EventID, opts *DispatchOptions) (*DispatchResponse, error)
}

type DispatchFunc func(ctx context.Context, event sdktypes.Event, opts *DispatchOptions) (*DispatchResponse, error)
