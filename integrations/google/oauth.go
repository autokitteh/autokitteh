package google

import (
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/google/internal/vars"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	// uiPath is the URL root path of a simple web UI to interact with users at
	// the beginning and the end of their 3-legged OAuth v2 flow with Google.
	uiPath = "/google/connect/"

	// oauthPath is the URL path for our handler to save new OAuth-based connections.
	oauthPath = "/google/oauth"
)

// handler is an autokitteh webhook which implements [http.Handler]
// to receive and dispatch asynchronous event notifications.
type handler struct {
	logger *zap.Logger
	vars   sdkservices.Vars
	oauth  sdkservices.OAuth
}

func NewHTTPHandler(l *zap.Logger, vars sdkservices.Vars, o sdkservices.OAuth) handler {
	return handler{logger: l, vars: vars, oauth: o}
}

// HandleOAuth receives an inbound redirect request from autokitteh's OAuth
// management service. This request contains an OAuth token (if the OAuth
// flow was successful) and form parameters for debugging and validation
// (either way). If all is well, it saves a new autokitteh connection.
// Either way, it redirects the user to success or failure webpages.
func (h handler) HandleOAuth(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("urlPath", r.URL.Path))

	// Handle errors (e.g. the user didn't authorize us) based on:
	// https://developers.google.com/identity/protocols/oauth2/web-server#handlingresponse
	e := r.FormValue("error")
	if e != "" {
		l.Warn("OAuth redirect request reported an error",
			zap.Error(errors.New(e)),
		)
		u := fmt.Sprintf("%serror.html?error=%s", uiPath, e)
		http.Redirect(w, r, u, http.StatusFound)
		return
	}

	rawOAuthData, _, err := sdkintegrations.GetOAuthDataFromURL(r.URL)
	if err != nil {
		l.Warn("Failed to decode OAuth data", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return

	}

	initData := sdktypes.EncodeVars(&vars.Vars{OAuthData: rawOAuthData})

	sdkintegrations.FinalizeConnectionInit(w, r, integrationID, initData)
}
