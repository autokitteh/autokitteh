package sdkintegrations

import (
	"context"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
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
	desc = kittehs.Must1(desc.Update(func(pb *sdktypes.IntegrationPB) {
		pb.Module = mod.Describe().ToProto()
	}))

	return &integration{desc: desc, mod: mod}
}

func (i *integration) Get() sdktypes.Integration { return i.desc }

func (i *integration) Configure(ctx context.Context, config string) (map[string]sdktypes.Value, error) {
	return i.mod.Configure(ctx, sdktypes.NewExecutorID(sdktypes.GetIntegrationID(i.desc)), config)
}

func (i *integration) Call(ctx context.Context, function sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	return i.mod.Call(ctx, function, args, kwargs)
}
