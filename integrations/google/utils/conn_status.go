package utils

import (
	"context"

	"go.autokitteh.dev/autokitteh/integrations/google/internal/vars"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// connStatus is an optional connection status check provided by the
// integration to AutoKitteh. The possible results are "init required"
// (the connection is not usable yet) and "using X" (where "X" is the
// authentication method: OAuth 2.0 (user), or JSON key (service account).
func ConnStatus(cvars sdkservices.Vars) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionStatus(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		if !cid.IsValid() {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "init required"), nil
		}

		vs, err := cvars.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			return sdktypes.InvalidStatus, err
		}

		if vs.Has(vars.OAuthData) {
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "using OAuth 2.0"), nil
		}
		if vs.Has(vars.JSON) {
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "using JSON key"), nil
		}

		return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "unrecognized auth"), nil
	})
}
