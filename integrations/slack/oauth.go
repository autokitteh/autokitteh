package slack

import (
	"errors"
	"net/http"
	"net/url"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/integrations/slack/api/auth"
	"go.autokitteh.dev/autokitteh/integrations/slack/api/bots"
	"go.autokitteh.dev/autokitteh/integrations/slack/internal/vars"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	// uiPath is the URL root path of a simple web UI to interact with
	// users at the beginning and the end of their 3-legged OAuth v2
	// flow to install a Slack app.
	uiPath = "/slack/connect/"

	// oauthPath is the URL path for our handler to save
	// new OAuth-based connections.
	oauthPath = "/slack/oauth"
)

// handler is an autokitteh webhook which implements [http.Handler]
// to receive and dispatch asynchronous event notifications.
type handler struct {
	logger *zap.Logger
	vars   sdkservices.Vars
}

func NewHandler(l *zap.Logger, sec sdkservices.Vars) http.Handler {
	return handler{logger: l, vars: sec}
}

// ServeHTTP receives an inbound redirect request from autokitteh's OAuth
// management service. This request contains an OAuth token (if the OAuth
// flow was successful) and form parameters for debugging and validation
// (either way). If all is well, it saves a new autokitteh connection.
// Either way, it redirects the user to success or failure webpages.
func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("urlPath", r.URL.Path))

	// Handle errors (e.g. the user didn't authorize us) based on:
	// https://developers.google.com/identity/protocols/oauth2/web-server#handlingresponse
	e := r.FormValue("error")
	if e != "" {
		l.Warn("OAuth redirect request reported an error",
			zap.Error(errors.New(e)),
		)
		u := uiPath + "error.html?error=" + url.QueryEscape(e)
		http.Redirect(w, r, u, http.StatusFound)
		return
	}

	oauthDataString, oauthData, err := sdkintegrations.GetOAuthDataFromURL(r.URL)
	if err != nil {
		l.Warn("Failed to decode OAuth data", zap.Error(err))
		http.Error(w, "Bad request: invalid OAuth data", http.StatusBadRequest)
		return
	}

	oauthToken := oauthData.Token
	if oauthToken == nil {
		l.Warn("OAuth data missing token")
		http.Error(w, "Bad request: missing OAuth token", http.StatusBadRequest)
		return
	}

	// Test the OAuth token's usability and get authoritative installation details.
	ctx := extrazap.AttachLoggerToContext(l, r.Context())
	authTest, err := auth.TestWithToken(ctx, oauthToken.AccessToken)
	if err != nil {
		e := "OAuth token test failed: " + err.Error()
		u := uiPath + "error.html?error=" + url.QueryEscape(e)
		http.Redirect(w, r, u, http.StatusFound)
		return
	}

	botInfo, err := bots.InfoWithToken(ctx, oauthToken.AccessToken, authTest)
	if err != nil {
		e := "Bot info request failed: " + err.Error()
		u := uiPath + "error.html?error=" + url.QueryEscape(e)
		http.Redirect(w, r, u, http.StatusFound)
		return
	}

	key := vars.KeyValue(botInfo.Bot.AppID, authTest.EnterpriseID, authTest.TeamID)
	initData := sdktypes.EncodeVars(
		vars.Vars{
			AppID:        botInfo.Bot.AppID,
			EnterpriseID: authTest.EnterpriseID,
			TeamID:       authTest.TeamID,
		},
	).
		Set(vars.KeyName, key, false).
		Set(vars.OAuthDataName, oauthDataString, true).
		Append(oauthData.ToVars()...)

	sdkintegrations.FinalizeConnectionInit(w, r, integrationID, initData)
}
