package salesforce

import (
	"errors"
	"net/http"

	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.uber.org/zap"
)

type handler struct {
	logger *zap.Logger
}

func NewHandler(l *zap.Logger) http.Handler {
	return handler{logger: l}
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, l := sdkintegrations.NewConnectionInit(h.logger, w, r, desc)

	l.Debug("Received Salesforce OAuth callback",
		zap.String("raw_query", r.URL.RawQuery),
		zap.String("state", r.FormValue("state")),
	)

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

	// TODO: Test the OAuth token's usability and get authoritative installation details.

	// TODO: Set connection vars.
	c.Finalize(data.ToVars().Set(OAuthDataName, raw, true))
}
