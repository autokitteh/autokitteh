package authz

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// id: resource id to check access to.
// action: action to check access for.
// data: optional data to check access with (such as list filter).
type CheckFunc = func(ctx context.Context, id sdktypes.ID, action string, opts ...CheckOpt) error
