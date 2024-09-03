package connections

import (
	"context"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/integrations/google/internal/vars"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// connStatus is an optional connection status check provided by
// the integration to AutoKitteh. The possible results are "init
// required" (the connection is not usable yet) and "using X".
func ConnStatus(cvars sdkservices.Vars) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionStatus(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		if !cid.IsValid() {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "init required"), nil
		}

		vs, err := cvars.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			return sdktypes.InvalidStatus, err
		}

		at := vs.Get(vars.AuthType)
		if !at.IsValid() || at.Value() == "" {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "init required"), nil
		}

		switch at.Value() {
		case integrations.JSONKey:
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "using JSON key"), nil
		case integrations.OAuth:
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "using OAuth 2.0"), nil
		default:
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "bad auth type"), nil
		}
	})
}
