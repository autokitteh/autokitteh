package chatgpt

import (
	"fmt"
	"net/http"
	"net/url"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	HeaderContentType = "Content-Type"
	ContentTypeForm   = "application/x-www-form-urlencoded"
)

var apiKeyVar = sdktypes.NewSymbol("api_key")

// handler is an autokitteh webhook which implements [http.Handler]
// to save data from web form submissions as connections.
type handler struct {
	logger *zap.Logger
}

func NewHTTPHandler(l *zap.Logger) http.Handler {
	return handler{logger: l}
}

// ServeHTTP saves a new autokitteh connection with user-submitted data.
func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		u := fmt.Sprintf("%serror.html?error=%s", uiPath, url.QueryEscape(e))
		http.Redirect(w, r, u, http.StatusFound)
		return
	}

	// Read and parse POST request body.
	err := r.ParseForm()
	if err != nil {
		l.Warn("Failed to parse inbound HTTP request",
			zap.Error(err),
		)
		e := "Form parsing error: " + err.Error()
		u := fmt.Sprintf("%serror.html?error=%s", uiPath, url.QueryEscape(e))
		http.Redirect(w, r, u, http.StatusFound)
		return
	}
	apiKey := r.Form.Get("key")

	sdkintegrations.FinalizeConnectionInit(w, r, integrationID, sdktypes.NewVars().Set(apiKeyVar, apiKey, true))
}
