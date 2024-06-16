package websockets

import (
	"fmt"
	"net/http"
	"net/url"
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
	// uiPath is the URL root path of a simple web UI to interact
	// with users to install a Slack Socket Mode app.
	uiPath = "/slack/connect"

	headerContentType = "Content-Type"
	contentTypeForm   = "application/x-www-form-urlencoded"
)

// HandleForm saves a new autokitteh connection, based on a user-submitted form.
func (h handler) HandleForm(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("urlPath", r.URL.Path))

	// Check the "Content-Type" header.
	contentType := r.Header.Get(headerContentType)
	if !strings.HasPrefix(contentType, contentTypeForm) {
		// This is probably an attack, so no user-friendliness.
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Read and parse POST request body.
	if err := r.ParseForm(); err != nil {
		l.Warn("Failed to parse inbound HTTP request", zap.Error(err))
		redirectToErrorPage(w, r, "form parsing error: "+err.Error())
		return
	}

	botToken := r.Form.Get("bot_token")
	appToken := r.Form.Get("app_token")

	// Test the Slack tokens usability and get authoritative installation details.
	ctx := extrazap.AttachLoggerToContext(l, r.Context())
	authTest, err := auth.TestWithToken(ctx, botToken)
	if err != nil {
		l.Warn("Slack OAuth token test failed", zap.Error(err))
		redirectToErrorPage(w, r, "token auth test failed: "+err.Error())
		return
	}

	botInfo, err := bots.InfoWithToken(ctx, botToken, authTest)
	if err != nil {
		l.Warn("Slack bot info request failed", zap.Error(err))
		redirectToErrorPage(w, r, "bot info request failed: "+err.Error())
		return
	}

	_, err = apps.ConnectionsOpenWithToken(ctx, h.vars, appToken)
	if err != nil {
		l.Warn("Slack websocket connection error", zap.Error(err))
		redirectToErrorPage(w, r, "websocket connection error: "+err.Error())
		return
	}

	initData := sdktypes.EncodeVars(
		vars.Vars{
			AppID:        botInfo.Bot.AppID,
			EnterpriseID: authTest.EnterpriseID,
			TeamID:       authTest.TeamID,
		},
	).
		Set(vars.AppTokenName, appToken, true).
		Set(vars.BotTokenName, botToken, true)

	// Open a new Socket Mode connection.
	h.OpenSocketModeConnection(botInfo.Bot.AppID, botToken, appToken)

	sdkintegrations.FinalizeConnectionInit(w, r, h.integrationID, initData)
}

func redirectToErrorPage(w http.ResponseWriter, r *http.Request, err string) {
	u := fmt.Sprintf("%s/error.html?error=%s", uiPath, url.QueryEscape(err))
	http.Redirect(w, r, u, http.StatusFound)
}
