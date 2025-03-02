package connections

import (
	"context"
	"io"
	"net/http"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/google/vars"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// connStatus is an optional connection status check provided by
// the integration to AutoKitteh. The possible results are "Init
// required" (the connection is not usable yet) and "Using X".
func ConnStatus(cvars sdkservices.Vars) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionStatus(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		if !cid.IsValid() {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		}

		vs, err := cvars.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			zap.L().Error("failed to read connection vars", zap.String("connection_id", cid.String()), zap.Error(err))
			return sdktypes.InvalidStatus, err
		}

		at := vs.Get(vars.AuthType)
		if !at.IsValid() || at.Value() == "" {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		}

		switch at.Value() {
		case integrations.JSONKey:
			if vs.GetValue(vars.JSON) == "" {
				return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
			}
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Using JSON key"), nil
		case integrations.OAuth:
			return common.CheckLegacyOAuthToken(vs)
		default:
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Bad auth type"), nil
		}
	})
}

// connTest is an optional connection test provided by the integration
// to AutoKitteh. It is used to verify that the connection is working
// as expected. The possible results are "OK" and "error".
func ConnTest(cvars sdkservices.Vars) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionTest(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		if !cid.IsValid() {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Init required"), nil
		}

		vs, err := cvars.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			zap.L().Error("failed to read connection vars", zap.String("connection_id", cid.String()), zap.Error(err))
			return sdktypes.InvalidStatus, err
		}

		at := vs.Get(vars.AuthType)
		if !at.IsValid() || at.Value() == "" {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		}

		var client *http.Client
		switch at.Value() {
		case "oauth":
			client, err = getOAuthClient(vs.GetValueByString("oauth_AccessToken"))
		case "jsonKey":
			client, err = getServiceAccountClient(vs.GetValue(vars.JSON))
		default:
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Unsupported auth type"), nil
		}
		if err != nil {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
		}

		// Make a simple API call to verify credentials.
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo?alt=json", nil)
		if err != nil {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
		}

		resp, err := client.Do(req)
		if err != nil {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return sdktypes.NewStatus(sdktypes.StatusCodeError, string(body)), nil
		}

		return sdktypes.NewStatus(sdktypes.StatusCodeOK, ""), nil
	})
}

func getOAuthClient(token string) (*http.Client, error) {
	t := &oauth2.Token{AccessToken: token, TokenType: "Bearer"}
	tokenSource := oauth2.StaticTokenSource(t)
	return oauth2.NewClient(context.Background(), tokenSource), nil
}

func getServiceAccountClient(json string) (*http.Client, error) {
	jwtConfig, err := google.JWTConfigFromJSON([]byte(json), "https://www.googleapis.com/auth/userinfo.email")
	if err != nil {
		return nil, err
	}

	return jwtConfig.Client(context.Background()), nil
}
