package telegram

import (
	"context"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	integrationName = "telegram"
)

var desc = common.Descriptor(integrationName, "Telegram", "/static/images/telegram.svg")

type integration struct{ vars sdkservices.Vars }

// connStatus is an optional connection status check provided by
// the integration to AutoKitteh. The possible results are "Init
// required" (the connection is not usable yet) and "Initialized".
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

		// Connection is valid if the secret token was saved.
		// This check is enough because the token will be saved only after successful authentication.
		at := vs.Get(SecretTokenVar)
		if at.Value() != "" {
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Initialized"), nil
		}

		return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
	})
}

// connTest is an optional connection test provided by the integration
// to AutoKitteh. It is used to verify that the connection is working
// as expected. The possible results are "OK" and "error".
func connTest(i *integration) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionTest(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		vs, errStatus, err := common.ReadVarsWithStatus(ctx, i.vars, cid)
		if errStatus.IsValid() || err != nil {
			return errStatus, err
		}

		botToken := vs.GetValue(BotTokenVar)

		_, err = getBotInfoWithToken(botToken, ctx)
		if err != nil {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Invalid bot token"), nil
		}

		return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Connection test successful"), nil
	})
}
