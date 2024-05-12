package websockets

import (
	"fmt"
	"net/http"
	"net/url"

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
	uiPath = "/slack/connect/"

	HeaderContentType = "Content-Type"
	ContentTypeForm   = "application/x-www-form-urlencoded"
)

// HandleForm saves a new autokitteh connection, based on a user-submitted form.
func (h handler) HandleForm(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("urlPath", r.URL.Path))

	// Check "Content-Type" header.
	ct := r.Header.Get(HeaderContentType)
	if ct != ContentTypeForm {
		l.Warn("Unexpected header value",
			zap.String("header", HeaderContentType),
			zap.String("got", ct),
			zap.String("want", ContentTypeForm),
		)
		e := fmt.Sprintf("Unexpected Content-Type header: %q", ct)
		u := uiPath + "error.html?error=" + url.QueryEscape(e)
		http.Redirect(w, r, u, http.StatusFound)
		return
	}

	// Read and parse POST request body.
	err := r.ParseForm()
	if err != nil {
		l.Warn("Failed to parse inbound HTTP request", zap.Error(err))
		e := "Form parsing error: " + err.Error()
		u := uiPath + "error.html?error=" + url.QueryEscape(e)
		http.Redirect(w, r, u, http.StatusFound)
		return
	}
	botToken := r.Form.Get("bot_token")
	appToken := r.Form.Get("app_token")

	// Test the Slack tokens usability and get authoritative installation details.
	ctx := extrazap.AttachLoggerToContext(l, r.Context())
	authTest, err := auth.TestWithToken(ctx, botToken)
	if err != nil {
		e := "Bot token test failed: " + err.Error()
		u := uiPath + "error.html?error=" + url.QueryEscape(e)
		http.Redirect(w, r, u, http.StatusFound)
		return
	}

	botInfo, err := bots.InfoWithToken(ctx, botToken, authTest)
	if err != nil {
		e := "Bot info request failed: " + err.Error()
		u := uiPath + "error.html?error=" + url.QueryEscape(e)
		http.Redirect(w, r, u, http.StatusFound)
		return
	}

	_, err = apps.ConnectionsOpenWithToken(ctx, h.vars, appToken)
	if err != nil {
		e := "Socket connection test failed: " + err.Error()
		u := uiPath + "error.html?error=" + url.QueryEscape(e)
		http.Redirect(w, r, u, http.StatusFound)
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
