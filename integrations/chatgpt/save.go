package chatgpt

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	headerContentType = "Content-Type"
	contentTypeForm   = "application/x-www-form-urlencoded"
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

	// Check the "Content-Type" header.
	contentType := r.Header.Get(headerContentType)
	if !strings.HasPrefix(contentType, contentTypeForm) {
		// Probably an attack, so no need for user-friendliness.
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Read and parse POST request body.
	if err := r.ParseForm(); err != nil {
		l.Warn("Failed to parse inbound HTTP request", zap.Error(err))
		redirectToErrorPage(w, r, "form parsing error: "+err.Error())
		return
	}
	apiKey := r.Form.Get("key")

	sdkintegrations.FinalizeConnectionInit(w, r, integrationID, sdktypes.NewVars().Set(apiKeyVar, apiKey, true))
}

func redirectToErrorPage(w http.ResponseWriter, r *http.Request, err string) {
	u := fmt.Sprintf("%s/error.html?error=%s", desc.ConnectionURL().Path, url.QueryEscape(err))
	http.Redirect(w, r, u, http.StatusFound)
}
