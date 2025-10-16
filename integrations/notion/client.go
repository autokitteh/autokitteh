package notion

import (
	"context"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/oauth"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var desc = common.Descriptor("notion", "Notion", "/static/images/notion.svg")

// connStatus is an optional connection status check provided by
// the integration to AutoKitteh. The possible results are "Init
// required" (the connection is not usable yet) and "Initialized".
func connStatus(cvars sdkservices.Vars) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionStatus(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		vs, errStatus, err := common.ReadVarsWithStatus(ctx, cvars, cid)
		if errStatus.IsValid() || err != nil {
			return errStatus, err
		}

		switch common.ReadAuthType(vs) {
		case "":
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		case integrations.OAuthDefault, integrations.OAuthPrivate:
			return common.CheckOAuthToken(vs)
		case integrations.APIKey:
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Using API key"), nil
		default:
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Bad auth type"), nil
		}
	})
}

// connTest is an optional connection test provided by the integration
// to AutoKitteh. It is used to verify that the connection is working
// as expected. The possible results are "OK" and "error".
func connTest(cvars sdkservices.Vars, o *oauth.OAuth) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionTest(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		l := zap.L().With(zap.String("connection_id", cid.String()))

		if !cid.IsValid() {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		}

		vs, err := cvars.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			l.Error("failed to get vars for connection "+cid.String()+": "+err.Error(), zap.Error(err))
			return sdktypes.InvalidStatus, err
		}

		authType := common.ReadAuthType(vs)
		switch authType {
		case "":
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		case integrations.OAuthDefault, integrations.OAuthPrivate:
			token := o.FreshToken(ctx, l, desc, vs)
			if token == nil {
				return sdktypes.NewStatus(sdktypes.StatusCodeError, "OAuth token not available"), nil
			}

			if err := validateAPIKey(ctx, token.AccessToken); err != nil {
				l.Debug("OAuth token validation failed for connection "+cid.String()+": "+err.Error(), zap.Error(err))
				return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
			}

		case integrations.APIKey:
			apiKey := vs.GetValue(common.ApiKeyVar)
			if apiKey == "" {
				return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "API key not configured"), nil
			}

			if err := validateAPIKey(ctx, apiKey); err != nil {
				l.Debug("API key validation failed for connection "+cid.String()+": "+err.Error(), zap.Error(err))
				return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
			}

		default:
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Bad auth type"), nil
		}

		return sdktypes.NewStatus(sdktypes.StatusCodeOK, ""), nil
	})
}
