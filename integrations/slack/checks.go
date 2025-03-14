package slack

import (
	"context"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/slack/api"
	"go.autokitteh.dev/autokitteh/integrations/slack/vars"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// status checks the connection's initialization status (is it
// initialized? what type of authentication is configured?). This
// ensures that the connection is at least theoretically usable.
func status(v sdkservices.Vars) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionStatus(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		vs, errStatus, err := common.ReadVarsWithStatus(ctx, v, cid)
		if errStatus.IsValid() || err != nil {
			return errStatus, err
		}

		switch common.ReadAuthType(vs) {
		case "":
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		// TODO(INT-267): Remove [integrations.OAuth] once the web UI is migrated too.
		case integrations.OAuth, integrations.OAuthDefault, integrations.OAuthPrivate:
			return common.CheckOAuthToken(vs)
		case integrations.SocketMode:
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Using Socket Mode"), nil
		default:
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Bad auth type"), nil
		}
	})
}

// test checks whether the connection is actually usable, i.e. the configured
// authentication credentials are valid and can be used to make API calls.
func test(v sdkservices.Vars) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionTest(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		vs, errStatus, err := common.ReadVarsWithStatus(ctx, v, cid)
		if errStatus.IsValid() || err != nil {
			return errStatus, err
		}

		var token string
		switch common.ReadAuthType(vs) {
		case "":
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		// TODO(INT-267): Remove [integrations.OAuth] once the web UI is migrated too.
		case integrations.OAuth, integrations.OAuthDefault, integrations.OAuthPrivate:
			token = vs.GetValue(common.OAuthAccessTokenVar)
		case integrations.SocketMode:
			token = vs.GetValue(vars.BotTokenVar)
		default:
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Bad auth type"), nil
		}

		if _, err = api.AuthTest(ctx, token); err != nil {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
		}

		return sdktypes.NewStatus(sdktypes.StatusCodeOK, ""), nil
	})
}
