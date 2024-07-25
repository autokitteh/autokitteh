package websockets

import (
	"net/http"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/integrations/slack/api/apps"
	"go.autokitteh.dev/autokitteh/integrations/slack/api/auth"
	"go.autokitteh.dev/autokitteh/integrations/slack/api/bots"
	"go.autokitteh.dev/autokitteh/integrations/slack/internal/vars"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	headerContentType = "Content-Type"
	contentTypeForm   = "application/x-www-form-urlencoded"
)

// HandleForm saves a new autokitteh connection, based on a user-submitted form.
func (h handler) HandleForm(w http.ResponseWriter, r *http.Request) {
	c, l := sdkintegrations.NewConnectionInit(h.logger, w, r, h.integration)

	// Check "Content-Type" header.
	contentType := r.Header.Get(headerContentType)
	if !strings.HasPrefix(contentType, contentTypeForm) {
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
	appToken := r.Form.Get("app_token")

	// Test the Slack tokens usability and get authoritative installation details.
	ctx := extrazap.AttachLoggerToContext(l, r.Context())
	authTest, err := auth.TestWithToken(ctx, botToken)
	if err != nil {
		l.Warn("Slack OAuth token test failed", zap.Error(err))
		c.AbortBadRequest("token auth test failed: " + err.Error())
		return
	}

	botInfo, err := bots.InfoWithToken(ctx, botToken, authTest)
	if err != nil {
		l.Warn("Slack bot info request failed", zap.Error(err))
		c.AbortBadRequest("bot info request failed: " + err.Error())
		return
	}

	_, err = apps.ConnectionsOpenWithToken(ctx, h.vars, appToken)
	if err != nil {
		l.Warn("Slack websocket connection error", zap.Error(err))
		c.AbortBadRequest("websocket connection error: " + err.Error())
		return
	}

	// Open a new Socket Mode connection.
	h.OpenSocketModeConnection(botInfo.Bot.AppID, botToken, appToken)

	c.Finalize(sdktypes.EncodeVars(vars.Vars{
		AppID:        botInfo.Bot.AppID,
		EnterpriseID: authTest.EnterpriseID,
		TeamID:       authTest.TeamID,
	}).
		Set(vars.AppTokenName, appToken, true).
		Set(vars.BotTokenName, botToken, true))
}
