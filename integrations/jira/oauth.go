package jira

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// handler is an autokitteh webhook which implements [http.Handler]
// to receive and dispatch asynchronous event notifications.
type handler struct {
	logger *zap.Logger
	oauth  sdkservices.OAuth
}

func NewHTTPHandler(l *zap.Logger, o sdkservices.OAuth) handler {
	return handler{logger: l, oauth: o}
}

// handleOAuth receives an inbound redirect request from autokitteh's OAuth
// management service. This request contains an OAuth token (if the OAuth
// flow was successful) and form parameters for debugging and validation
// (either way). If all is well, it saves a new autokitteh connection.
// Either way, it redirects the user to success or failure webpages.
func (h handler) handleOAuth(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("urlPath", r.URL.Path))

	// Handle errors (e.g. the user didn't authorize us) based on:
	// https://developers.google.com/identity/protocols/oauth2/web-server#handlingresponse
	e := r.FormValue("error")
	if e != "" {
		l.Warn("OAuth redirect request reported an error", zap.Error(errors.New(e)))
		u := fmt.Sprintf("%serror.html?error=%s", uiPath, url.QueryEscape(e))
		http.Redirect(w, r, u, http.StatusFound)
		return
	}

	_, data, err := sdkintegrations.GetOAuthDataFromURL(r.URL)
	if err != nil {
		l.Warn("OAuth redirect request with invalid data parameter", zap.Error(err))
		u := uiPath + "error.html?error=" + url.QueryEscape("invalid data parameter")
		http.Redirect(w, r, u, http.StatusFound)
		return
	}

	oauthToken := data.Token
	if oauthToken == nil {
		l.Warn("OAuth redirect request without token in data parameter", zap.Error(err))
		u := uiPath + "error.html?error=" + url.QueryEscape("missing OAuth token")
		http.Redirect(w, r, u, http.StatusFound)
		return
	}

	// TODO(ENG-965):
	// Create a webhook to receive, parse, and dispatch Jira events,
	// and retrieve authoritative app and installation details.

	initData := sdktypes.NewVars(data.ToVars()...)

	sdkintegrations.FinalizeConnectionInit(w, r, integrationID, initData)
}
