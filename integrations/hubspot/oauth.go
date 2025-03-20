package hubspot

import (
	"errors"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/oauth"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// handler is an autokitteh webhook which implements [http.Handler]
// to receive and dispatch asynchronous event notifications.
type handler struct {
	logger *zap.Logger
	oauth  *oauth.OAuth
}

func NewHTTPHandler(l *zap.Logger, o *oauth.OAuth) http.Handler {
	return handler{logger: l, oauth: o}
}

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

	_, data, err := sdkintegrations.GetOAuthDataFromURL(r.URL)
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

	// Test the OAuth token's usability and get authoritative connection details.
	url := "https://api.hubapi.com/crm/v3/owners/"
	auth := "Bearer " + oauthToken.AccessToken
	if _, err := common.HTTPGet(r.Context(), url, auth); err != nil {
		l.Warn("failed to test HubSpot OAuth token", zap.Error(err))
		c.AbortServerError("failed to test OAuth token")
		return
	}

	c.Finalize(sdktypes.NewVars(data.ToVars()...))
}
