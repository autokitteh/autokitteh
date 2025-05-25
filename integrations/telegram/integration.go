package telegram

import (
	"context"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/telegram/api"
	"go.autokitteh.dev/autokitteh/integrations/telegram/vars"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
    integrationName = "telegram"
)

var (
    integrationID = sdktypes.NewIntegrationIDFromName(integrationName)
    desc = common.Descriptor(integrationName, "Telegram", "/static/images/telegram.svg")
)

type integration struct {
    vars sdkservices.Vars
}

func New(cvars sdkservices.Vars) sdkservices.Integration {
    i := &integration{vars: cvars}
    return sdkintegrations.NewIntegration(
        desc, sdkmodule.New(), connStatus(i), connTest(i),
        sdkintegrations.WithConnectionConfigFromVars(cvars),
    )
}

// connStatus is an optional connection status check provided by
// the integration to AutoKitteh. The possible results are "Init
// required" (the connection is not usable yet) and "Using X".
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

        botToken := vs.GetValue(vars.BotTokenVar)
        if botToken == "" {
            return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
        }

        // Check if we have bot info stored
        botUsername := vs.GetValue(vars.BotUsernameVar)
        if botUsername != "" {
            return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Using Telegram bot: @"+botUsername), nil
        }

        return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Using Telegram bot"), nil
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

        botToken := vs.GetValue(vars.BotTokenVar)
        if botToken == "" {
            return sdktypes.NewStatus(sdktypes.StatusCodeError, "Bot token not configured"), nil
        }

        // Test the connection by calling getMe
        client := api.NewClient(botToken)
        user, err := client.GetMe(ctx)
        if err != nil {
            return sdktypes.NewStatus(sdktypes.StatusCodeError, "Failed to connect to Telegram: "+err.Error()), nil
        }

        if !user.IsBot {
            return sdktypes.NewStatus(sdktypes.StatusCodeError, "Token does not belong to a bot"), nil
        }

        return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Connection successful - bot: @"+user.Username), nil
    })
}