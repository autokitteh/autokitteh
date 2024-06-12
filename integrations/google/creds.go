package google

import (
	"net/http"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/google/internal/vars"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	// credsPath is the URL path for our handler to save a new autokitteh
	// credentials-based connection, after the user submits it via a web form.
	credsPath = "/google/save"

	headerContentType = "Content-Type"
	contentTypeForm   = "application/x-www-form-urlencoded"
)

// HandleCreds saves a new autokitteh connection with a user-submitted token.
func (h handler) HandleCreds(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("urlPath", r.URL.Path))

	// Check the "Content-Type" header.
	contentType := r.Header.Get(headerContentType)
	if !strings.HasPrefix(contentType, contentTypeForm) {
		// This is probably an attack, so no user-friendliness.
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Read and parse POST request body.
	err := r.ParseForm()
	if err != nil {
		l.Warn("Failed to parse inbound HTTP request", zap.Error(err))
		redirectToErrorPage(w, r, "form parsing error: "+err.Error())
		return
	}

	initData := sdktypes.EncodeVars(&vars.Vars{JSON: r.Form.Get("json")})

	sdkintegrations.FinalizeConnectionInit(w, r, integrationID, initData)
}
