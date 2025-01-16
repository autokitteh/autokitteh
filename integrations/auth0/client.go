package auth0

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type integration struct{ vars sdkservices.Vars }

var (
	integrationID = sdktypes.NewIntegrationIDFromName("auth0")

	authType         = sdktypes.NewSymbol("auth_type")
	clientIDName     = sdktypes.NewSymbol("client_id")
	clientSecretName = sdktypes.NewSymbol("client_secret")
	domainName       = sdktypes.NewSymbol("auth0_domain")
	authToken        = sdktypes.NewSymbol("oauth_AccessToken")
)

var desc = kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
	IntegrationId: integrationID.String(),
	UniqueName:    "auth0",
	DisplayName:   "Auth0",
	Description:   "Auth0 is an identity platform that provides authentication and authorization services.",
	LogoUrl:       "/static/images/auth0.svg",
	UserLinks: map[string]string{
		"1 Auth0 API reference": "https://auth0.com/docs/api/management/v2",
		"2 Python client API":   "https://auth0-python.readthedocs.io/en/latest/",
	},
	ConnectionUrl: "/auth0/connect",
	ConnectionCapabilities: &sdktypes.ConnectionCapabilitiesPB{
		RequiresConnectionInit: true,
	},
}))

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

		at := vs.Get(authType)
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
// as expected. The possible results are "OK" and "error
// https://auth0.com/docs/api/management/v2/stats/get-active-users
func connTest(i *integration) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionTest(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		if !cid.IsValid() {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Init required"), nil
		}

		vs, err := i.vars.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			return sdktypes.InvalidStatus, err
		}

		token := vs.Get(authToken).Value()
		domain := vs.Get(domainName).Value()
		url := fmt.Sprintf("https://%s/api/v2/stats/active-users", domain)

		req, err := http.NewRequest(http.MethodGet, url, nil)
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
