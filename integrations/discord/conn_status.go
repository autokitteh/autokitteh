package discord

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// connStatus is an optional connection status check provided by the
// integration to AutoKitteh. The possible results are "init required"
// (the connection is not usable yet) and "using X" (where "X" is the
// authentication method: bot token.
func connStatus(cvars sdkservices.Vars) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionStatus(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		if !cid.IsValid() {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "init required"), nil
		}

		vs, err := cvars.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			return sdktypes.InvalidStatus, err
		}

		if vs.Has(botToken) {
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "using bot token"), nil
		}

		return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "unrecognized auth"), nil
	})
}
