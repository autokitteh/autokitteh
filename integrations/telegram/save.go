package telegram

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/telegram/api"
	"go.autokitteh.dev/autokitteh/integrations/telegram/vars"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// handler is an autokitteh webhook which implements [http.Handler]
// to save data from web form submissions as connections.
type handler struct {
    logger *zap.Logger
    vars   sdkservices.Vars
}

func NewHTTPHandler(l *zap.Logger, v sdkservices.Vars) http.Handler {
    return handler{logger: l, vars: v}
}

// ServeHTTP saves a new autokitteh connection with user-submitted data.
func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    c, l := sdkintegrations.NewConnectionInit(h.logger, w, r, desc)

    // Check the "Content-Type" header.
    if common.PostWithoutFormContentType(r) {
        ct := r.Header.Get(common.HeaderContentType)
        l.Warn("save connection: unexpected POST content type", zap.String("content_type", ct))
        c.AbortBadRequest("unexpected content type")
        return
    }

    // Read and parse POST request body.
    if err := r.ParseForm(); err != nil {
        l.Warn("Failed to parse incoming HTTP request", zap.Error(err))
        c.AbortBadRequest("form parsing error")
        return
    }

    botToken := r.Form.Get("bot_token")
    if botToken == "" {
        l.Warn("Bot token not provided in request form")
        c.AbortBadRequest("bot token is missing")
        return
    }

    // Validate the bot token by calling getMe
    client := api.NewClient(botToken)
    user, err := client.GetMe(r.Context())
    if err != nil {
        l.Warn("Failed to validate bot token", zap.Error(err))
        c.AbortBadRequest("failed to validate bot token")
        return
    }

    if !user.IsBot {
        l.Warn("Token does not belong to a bot")
        c.AbortBadRequest("token does not belong to a bot")
        return
    }

    // Create variables to save
    varsToSave := sdktypes.NewVars().
        Set(vars.BotTokenVar, botToken, true).
        Set(vars.BotIDVar, fmt.Sprintf("%d", user.ID), false).
        Set(vars.BotUsernameVar, user.Username, false).
        Set(vars.BotFirstNameVar, user.FirstName, false)

    // Add optional fields if they exist
    if user.LastName != "" {
        varsToSave = varsToSave.Set(vars.BotLastNameVar, user.LastName, false)
    }

    // Add additional bot capabilities if available
    if user.CanJoinGroups {
        varsToSave = varsToSave.Set(vars.BotCanJoinGroupsVar, "true", false)
    }
    if user.CanReadAllGroupMessages {
        varsToSave = varsToSave.Set(vars.BotCanReadAllGroupMessagesVar, "true", false)
    }
    if user.SupportsInlineQueries {
        varsToSave = varsToSave.Set(vars.BotSupportsInlineQueriesVar, "true", false)
    }

    // Add webhook secret if provided
    webhookSecret := r.Form.Get("webhook_secret")
    if webhookSecret != "" {
        varsToSave = varsToSave.Set(vars.WebhookSecretVar, webhookSecret, true)
    }

    c.Finalize(varsToSave)
}