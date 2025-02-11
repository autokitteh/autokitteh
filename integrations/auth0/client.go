package auth0

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	desc = common.Descriptor("auth0", "Auth0", "/static/images/auth0.svg")

	authTypeVar         = sdktypes.NewSymbol("auth_type")
	clientIDNameVar     = sdktypes.NewSymbol("client_id")
	clientSecretNameVar = sdktypes.NewSymbol("client_secret")
	domainNameVar       = sdktypes.NewSymbol("auth0_domain")
	authTokenNameVar    = sdktypes.NewSymbol("oauth_AccessToken")
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
			return sdktypes.InvalidStatus, err
		}

		// TODO(INT-124): Use the refresh token to get a new access token.
		token := vs.Get(authTokenNameVar).Value()
		domain := vs.Get(domainNameVar).Value()
		// https://auth0.com/docs/api/management/v2/stats/get-active-users
		url := fmt.Sprintf("https://%s/api/v2/stats/active-users", domain)

		timeoutCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(timeoutCtx, http.MethodGet, url, nil)
		if err != nil {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
		}
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Invalid OAuth token"), nil
		}

		return sdktypes.NewStatus(sdktypes.StatusCodeOK, "OK"), nil
	})
}
