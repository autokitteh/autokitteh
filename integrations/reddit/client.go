package reddit

import (
	"context"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/integrations/github/vars"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func New(cvars sdkservices.Vars) sdkservices.Integration {
	return sdkintegrations.NewIntegration(
		desc,
		sdkmodule.New(),
		connStatus(cvars),
		connTest(cvars),
		sdkintegrations.WithConnectionConfigFromVars(cvars),
	)
}

// connStatus is an optional connection status check provided by
// the integration to AutoKitteh. The possible results are "Init
// required" (the connection is not usable yet) and "Initialized".
func connStatus(cvars sdkservices.Vars) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionStatus(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		if !cid.IsValid() {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		}

		vs, err := cvars.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			return sdktypes.InvalidStatus, err
		}

		at := vs.Get(vars.AuthType)
		if !at.IsValid() || at.Value() == "" {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		}

		if at.Value() == integrations.OAuthPrivate {
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Initialized"), nil
		}
		return sdktypes.NewStatus(sdktypes.StatusCodeError, "Bad auth type"), nil
	})
}

// connTest is an optional connection test provided by the integration
// to AutoKitteh. It is used to verify that the connection is working
// as expected. The possible results are "OK" and "error".
func connTest(cvars sdkservices.Vars) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionTest(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		l := zap.L().With(zap.String("connection_id", cid.String()))

		if !cid.IsValid() {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		}

		// Retrieve connection variables.
		vs, err := cvars.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			l.Error("Failed to get connection variables for connection "+cid.String()+": "+err.Error(), zap.Error(err))
			return sdktypes.InvalidStatus, err
		}

		// Get credentials from variables
		clientID := vs.GetValue(clientIDVar)
		clientSecret := vs.GetValue(clientSecretVar)
		userAgent := vs.GetValue(userAgentVar)
		username := vs.GetValue(usernameVar)
		password := vs.GetValue(passwordVar)

		// Check if required credentials are present
		if clientID == "" || clientSecret == "" || userAgent == "" {
			l.Debug("Missing required credentials")
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		}

		// Validate credentials by attempting to authenticate with Reddit API
		if err := validateRedditCredentials(ctx, clientID, clientSecret, username, password); err != nil {
			l.Debug("Credential validation failed for connection "+cid.String()+": "+err.Error(), zap.Error(err))
			return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
		}

		return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Connection validated successfully"), nil
	})
}
