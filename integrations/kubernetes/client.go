package kubernetes

import (
	"context"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	desc          = common.Descriptor("kubernetes", "Kubernetes", "/static/images/k8s.svg")
	configFileVar = sdktypes.NewSymbol("configFile")
	authTypeVar   = sdktypes.NewSymbol("authType")
)

type integration struct{ vars sdkservices.Vars }

func New(vars sdkservices.Vars) sdkservices.Integration {
	i := &integration{vars: vars}
	return sdkintegrations.NewIntegration(
		desc,
		sdkmodule.New(),
		connStatus(i),
		connTest(i),
		sdkintegrations.WithConnectionConfigFromVars(vars),
	)
}

// TODO: Implement the connection status functions for Kubernetes integration.
func connStatus(_ *integration) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionStatus(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
	})
}

// TODO: Implement the connection test function for Kubernetes integration.
func connTest(_ *integration) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionTest(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		return sdktypes.NewStatus(sdktypes.StatusCodeUnspecified, "Not implemented"), nil
	})
}
