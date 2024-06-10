package sdkintegrations

import (
	"context"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type integration struct {
	desc       sdktypes.Integration
	mod        sdkmodule.Module
	connTest   func(context.Context, sdktypes.ConnectionID) (sdktypes.Status, error)
	connStatus func(context.Context, sdktypes.ConnectionID) (sdktypes.Status, error)
	connConfig func(context.Context, sdktypes.ConnectionID) (map[string]string, error)
}

type OptFn func(*integration)

func WithConnectionTest(fn func(context.Context, sdktypes.ConnectionID) (sdktypes.Status, error)) OptFn {
	return func(i *integration) { i.connTest = fn }
}

func WithConnectionStatus(fn func(context.Context, sdktypes.ConnectionID) (sdktypes.Status, error)) OptFn {
	return func(i *integration) { i.connStatus = fn }
}

func WithConnectionConfig(fn func(context.Context, sdktypes.ConnectionID) (map[string]string, error)) OptFn {
	return func(i *integration) { i.connConfig = fn }
}

func WithConnectionConfigFromVars(cvars sdkservices.Vars) OptFn {
	if cvars == nil {
		return WithoutConnectionConfig()
	}

	return func(i *integration) {
		i.connConfig = func(ctx context.Context, cid sdktypes.ConnectionID) (map[string]string, error) {
			vs, err := cvars.Reveal(ctx, sdktypes.NewVarScopeID(cid))
			if err != nil {
				return nil, err
			}

			return kittehs.ListToMap(vs, func(v sdktypes.Var) (string, string) {
				return v.Name().String(), v.Value()
			}), nil
		}
	}
}

func WithoutConnectionConfig() OptFn {
	return func(i *integration) {
		i.connConfig = func(context.Context, sdktypes.ConnectionID) (map[string]string, error) {
			return nil, nil
		}
	}
}

// NewIntegration creates a new integration, augmenting the given `desc` with
// the members defintion from `mod`.
func NewIntegration(
	desc sdktypes.Integration,
	mod sdkmodule.Module,
	opts ...OptFn,
) sdkservices.Integration {
	i := &integration{mod: mod}

	for _, opt := range opts {
		opt(i)
	}

	i.desc = desc.UpdateModule(mod.Describe())

	if i.connTest != nil {
		i.desc = i.desc.WithConnectionCapabilities(i.desc.ConnectionCapabilities().WithSupportsConnectionTest(true))
	}

	connURL := i.desc.ConnectionURL()

	i.desc = i.desc.WithConnectionCapabilities(i.desc.ConnectionCapabilities().WithSupportsConnectionInit(connURL != nil && connURL.String() != ""))

	if i.connStatus == nil {
		i.connStatus = func(_ context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
			if !cid.IsValid() {
				return i.desc.InitialConnectionStatus(), nil
			}

			return sdktypes.NewStatus(sdktypes.StatusCodeUnspecified, ""), nil
		}
	}

	return i
}

func (i *integration) Get() sdktypes.Integration { return i.desc }

func (i *integration) Configure(ctx context.Context, cid sdktypes.ConnectionID) (vs map[string]sdktypes.Value, cfg map[string]string, err error) {
	if vs, err = i.mod.Configure(ctx, sdktypes.NewExecutorID(i.desc.ID()), cid); err != nil {
		return
	}

	cfg, err = i.GetConnectionConfig(ctx, cid)

	return
}

func (i *integration) Call(ctx context.Context, function sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	return i.mod.Call(ctx, function, args, kwargs)
}

func (i *integration) TestConnection(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
	if i.connTest == nil {
		return sdktypes.InvalidStatus, sdkerrors.ErrNotImplemented
	}

	return i.connTest(ctx, cid)
}

func (i *integration) GetConnectionStatus(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
	if i.connStatus == nil {
		return sdktypes.InvalidStatus, sdkerrors.ErrNotImplemented
	}

	return i.connStatus(ctx, cid)
}

func (i *integration) GetConnectionConfig(ctx context.Context, cid sdktypes.ConnectionID) (map[string]string, error) {
	if i.connConfig == nil {
		return nil, sdkerrors.ErrNotImplemented
	}

	return i.connConfig(ctx, cid)
}
