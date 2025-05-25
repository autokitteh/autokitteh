package telegram

import (
	"context"
	"fmt"

	"go.autokitteh.dev/autokitteh/integrations/telegram/api"
	"go.autokitteh.dev/autokitteh/integrations/telegram/vars"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// status returns a function that checks the status of a Telegram connection
func status(v sdkservices.Vars) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionStatus(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		if !cid.IsValid() {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Connection ID is invalid"), nil
		}

		vs, err := v.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, fmt.Sprintf("Failed to get connection variables: %v", err)), nil
		}

		token := vs.GetValue(vars.BotTokenVar)
		if token == "" {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Bot token is not configured"), nil
		}

		// Try to get bot information to verify the token
		client := api.NewClient(token)
		_, err = client.GetMe(ctx)
		if err != nil {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, fmt.Sprintf("Failed to verify bot token: %v", err)), nil
		}

		return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Connection is active"), nil
	})
}

// test returns a function that tests a Telegram connection
func test(v sdkservices.Vars) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionTest(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		if !cid.IsValid() {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Connection ID is invalid"), nil
		}

		vs, err := v.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, fmt.Sprintf("Failed to get connection variables: %v", err)), nil
		}

		token := vs.GetValue(vars.BotTokenVar)
		if token == "" {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Bot token is not configured"), nil
		}

		// Test the connection by calling getMe
		client := api.NewClient(token)
		user, err := client.GetMe(ctx)
		if err != nil {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, fmt.Sprintf("Failed to connect to Telegram API: %v", err)), nil
		}

		if !user.IsBot {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Token does not belong to a bot"), nil
		}

		return sdktypes.NewStatus(sdktypes.StatusCodeOK, fmt.Sprintf("Successfully connected to bot: %s (@%s)", user.FirstName, user.Username)), nil
	})
}
