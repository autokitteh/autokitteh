package sdkservices

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type DispatchOptions struct {
	// If set, dispatch only to deployments in this project. Can be either n project ID or name.
	Project string

	// If set, dispatch only to this specific deployment.
	DeploymentID sdktypes.DeploymentID
}

type Dispatcher interface {
	Dispatch(ctx context.Context, event sdktypes.Event, opts *DispatchOptions) (sdktypes.EventID, error)
	Redispatch(ctx context.Context, eventID sdktypes.EventID, opts *DispatchOptions) (sdktypes.EventID, error)
}
