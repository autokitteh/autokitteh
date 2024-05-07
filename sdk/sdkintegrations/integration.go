package sdkintegrations

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type integration struct {
	desc sdktypes.Integration
	mod  sdkmodule.Module
}

// NewIntegration creates a new integration, augmenting the given `desc` with
// the members defintion from `mod`.
func NewIntegration(desc sdktypes.Integration, mod sdkmodule.Module) sdkservices.Integration {
	return &integration{desc: desc.UpdateModule(mod.Describe()), mod: mod}
}

func (i *integration) Get() sdktypes.Integration { return i.desc }

func (i *integration) Configure(ctx context.Context, cid sdktypes.ConnectionID) (map[string]sdktypes.Value, error) {
	return i.mod.Configure(ctx, sdktypes.NewExecutorID(i.desc.ID()), cid)
}

func (i *integration) Call(ctx context.Context, function sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	return i.mod.Call(ctx, function, args, kwargs)
}
