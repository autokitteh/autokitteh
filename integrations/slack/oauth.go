package slack

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/integrations/slack/api/auth"
	"go.autokitteh.dev/autokitteh/integrations/slack/api/bots"
	"go.autokitteh.dev/autokitteh/integrations/slack/internal/vars"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// handler is an autokitteh webhook which implements [http.Handler]
// to receive and dispatch asynchronous event notifications.
type handler struct {
	logger *zap.Logger
}

func NewHandler(l *zap.Logger) http.Handler {
	return handler{logger: l}
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
		l.Warn("OAuth redirect request reported an error", zap.Error(errors.New(e)))
		redirectToErrorPage(w, r, e)
		return
	}

	raw, data, err := sdkintegrations.GetOAuthDataFromURL(r.URL)
	if err != nil {
		l.Warn("Invalid data in OAuth redirect request", zap.Error(err))
		redirectToErrorPage(w, r, "invalid data parameter")
		return
	}

	oauthToken := data.Token
	if oauthToken == nil {
		l.Warn("Missing token in OAuth redirect request", zap.Any("data", data))
		redirectToErrorPage(w, r, "missing OAuth token")
		return
	}

	// Test the OAuth token's usability and get authoritative installation details.
	ctx := extrazap.AttachLoggerToContext(l, r.Context())
	authTest, err := auth.TestWithToken(ctx, oauthToken.AccessToken)
	if err != nil {
		l.Warn("Slack OAuth token test failed", zap.Error(err))
		redirectToErrorPage(w, r, "token auth test failed: "+err.Error())
		return
	}

	botInfo, err := bots.InfoWithToken(ctx, oauthToken.AccessToken, authTest)
	if err != nil {
		l.Warn("Slack bot info request failed", zap.Error(err))
		redirectToErrorPage(w, r, "bot info request failed: "+err.Error())
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
		Set(vars.OAuthDataName, raw, true).
		Append(data.ToVars()...)

	sdkintegrations.FinalizeConnectionInit(w, r, integrationID, initData)
}

func redirectToErrorPage(w http.ResponseWriter, r *http.Request, err string) {
	u := fmt.Sprintf("%s/error.html?error=%s", desc.ConnectionURL().Path, url.QueryEscape(err))
	http.Redirect(w, r, u, http.StatusFound)
}
