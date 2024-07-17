package discord

import (
	// "context"
	"net/http"

	"go.uber.org/zap"
	// "golang.org/x/oauth2"

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

// type vars struct {
// 	OAuthData string `var:"secret"`
// }

func NewHTTPHandler(l *zap.Logger, o sdkservices.OAuth, v sdkservices.Vars) handler {
	return handler{logger: l, oauth: o, vars: v}
}

// handleOAuth receives an inbound redirect request from autokitteh's OAuth
// management service. This request contains an OAuth token (if the OAuth
// flow was successful) and form parameters for debugging and validation
// (either way). If all is well, it saves a new autokitteh connection.
// Either way, it redirects the user to success or failure webpages.
func (h handler) handleOAuth(w http.ResponseWriter, r *http.Request) {
	c, l := sdkintegrations.NewConnectionInit(h.logger, w, r, desc)

	// Handle errors (e.g. the user didn't authorize us) based on:
	// https://developers.google.com/identity/protocols/oauth2/web-server#handlingresponse
	// e := r.FormValue("error")
	// if e != "" {
	// 	l.Warn("OAuth redirect request reported an error", zap.Error(errors.New(e)))
	// 	c.Abort(e)
	// 	return
	// }

	raw, data, err := sdkintegrations.GetOAuthDataFromURL(r.URL)
	if err != nil {
		l.Warn("Invalid data in OAuth redirect request", zap.Error(err))
		c.Abort("invalid data parameter")
		return

	}

	oauthToken := data.Token
	if oauthToken == nil {
		l.Warn("Missing token in OAuth redirect request", zap.Any("data", data))
		c.Abort("missing OAuth token")
		return
	}

	// Encoding "OAuthData" and "JSON", but not "FormID", so we don't overwrite
	// the value that was already written there by the creds.go passthrough.
	c.Finalize(sdktypes.NewVars(sdktypes.NewVar(sdktypes.NewSymbol(oauthToken.TokenType), raw, true)))
}

// func (h handler) tokenSource(ctx context.Context, t *oauth2.Token) oauth2.TokenSource {
// 	cfg, _, err := h.oauth.Get(ctx, "google")
// 	if err != nil {
// 		return nil
// 	}
// 	return cfg.TokenSource(ctx, t)
// }
