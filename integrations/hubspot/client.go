package hubspot

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	desc = common.LegacyDescriptor("hubspot", "HubSpot", "/static/images/hubspot.svg")

	authTypeVar = sdktypes.NewSymbol("auth_type")
)

type integration struct{ vars sdkservices.Vars }

func New(cvars sdkservices.Vars) sdkservices.Integration {
	i := &integration{vars: cvars}
	return sdkintegrations.NewIntegration(
		desc,
		sdkmodule.New(),
		connStatus(i),
		connTest(i),
		sdkintegrations.WithConnectionConfigFromVars(cvars),
	)
}

func connStatus(i *integration) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionStatus(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		if !cid.IsValid() {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		}

		vs, err := i.vars.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			zap.L().Error("failed to read connection vars", zap.String("connection_id", cid.String()), zap.Error(err))
			return sdktypes.InvalidStatus, err
		}

		at := vs.Get(authTypeVar)
		if !at.IsValid() || at.Value() == "" {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		}

		if at.Value() == integrations.OAuth {
			token := vs.GetValueByString("oauth_AccessToken")
			if token == "" {
				return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
			}
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Using OAuth 2.0"), nil
		}

		return sdktypes.NewStatus(sdktypes.StatusCodeError, "Bad auth type"), nil
	})
}

// connTest is an optional connection test provided by the integration
// to AutoKitteh. It is used to verify that the connection is working
// as expected. The possible results are "OK" and "error".
func connTest(i *integration) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionTest(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		if !cid.IsValid() {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Init required"), nil
		}

		vs, err := i.vars.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			zap.L().Error("failed to read connection vars", zap.String("connection_id", cid.String()), zap.Error(err))
			return sdktypes.InvalidStatus, err
		}

		token := vs.GetValueByString("oauth_RefreshToken")
		url := "https://api.hubapi.com/oauth/v1/refresh-tokens/" + token
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Bad OAuth refresh token"), nil
		}

		return sdktypes.NewStatus(sdktypes.StatusCodeOK, "OK"), nil
	})
}
