package sdkexecutor

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Function = func(context.Context, []sdktypes.Value, map[string]sdktypes.Value) (sdktypes.Value, error)

type Caller interface {
	Call(context.Context, sdktypes.Value, []sdktypes.Value, map[string]sdktypes.Value) (sdktypes.Value, error)
}
