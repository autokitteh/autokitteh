package hubspot

import (
	"errors"
	"net/http"

	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.uber.org/zap"
)

// handler is an autokitteh webhook which implements [http.Handler]
// to receive and dispatch asynchronous event notifications.
type handler struct {
	logger *zap.Logger
	oauth  sdkservices.OAuth
}

func NewHTTPHandler(l *zap.Logger, o sdkservices.OAuth) http.Handler {
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

	// TODO: Test the OAuth token's usability and get authoritative installation details.

	c.Finalize(sdktypes.NewVars(data.ToVars()...))
}
