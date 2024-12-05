package slack

import (
	"errors"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/integrations/slack/api/auth"
	"go.autokitteh.dev/autokitteh/integrations/slack/api/bots"
	"go.autokitteh.dev/autokitteh/integrations/slack/internal/vars"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// handler is an autokitteh webhook which implements [http.Handler]
// to receive and dispatch asynchronous event notifications.
type handler struct {
	logger *zap.Logger
	oauth  sdkservices.OAuth
	vars   sdkservices.Vars
}

func NewHandler(l *zap.Logger, o sdkservices.OAuth, v sdkservices.Vars) handler {
	return handler{logger: l, oauth: o, vars: v}
}

// ServeHTTP receives an inbound redirect request from autokitteh's OAuth
// management service. This request contains an OAuth token (if the OAuth
// flow was successful) and form parameters for debugging and validation
// (either way). If all is well, it saves a new autokitteh connection.
// Either way, it redirects the user to success or failure webpages.
func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, l := sdkintegrations.NewConnectionInit(h.logger, w, r, desc)

	// Handle errors (e.g. the user didn't authorize us) based on:
	// https://developers.google.com/identity/protocols/oauth2/web-server#handlingresponse
	e := r.FormValue("error")
	if e != "" {
		l.Warn("OAuth redirect request reported an error", zap.Error(errors.New(e)))
		c.AbortBadRequest(e)
		return
	}

	raw, data, err := sdkintegrations.GetOAuthDataFromURL(r.URL)
	if err != nil {
		l.Warn("Invalid data in OAuth redirect request", zap.Error(err))
		c.AbortBadRequest("invalid data parameter")
		return
	}

	oauthToken := data.Token
	if oauthToken == nil {
		l.Warn("Missing token in OAuth redirect request", zap.Any("data", data))
		c.AbortBadRequest("missing OAuth token")
		return
	}

	// Test the OAuth token's usability and get authoritative installation details.
	ctx := extrazap.AttachLoggerToContext(l, r.Context())
	authTest, err := auth.TestWithToken(ctx, oauthToken.AccessToken)
	if err != nil {
		l.Warn("Slack OAuth token test failed", zap.Error(err))
		e := "token auth test failed: " + err.Error()
		c.AbortBadRequest(e)
		return
	}

	botInfo, err := bots.InfoWithToken(ctx, oauthToken.AccessToken, authTest)
	if err != nil {
		l.Warn("Slack bot info request failed", zap.Error(err))
		e := "bot info request failed: " + err.Error()
		c.AbortBadRequest(e)
		return
	}

	key := vars.KeyValue(botInfo.Bot.AppID, authTest.EnterpriseID, authTest.TeamID)
	c.Finalize(sdktypes.EncodeVars(vars.Vars{
		AppID:        botInfo.Bot.AppID,
		EnterpriseID: authTest.EnterpriseID,
		TeamID:       authTest.TeamID,
	}).
		Set(vars.KeyName, key, false).
		Set(vars.OAuthDataName, raw, true).
		Append(data.ToVars()...))
}
