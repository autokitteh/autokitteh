package jira

import (
	"net/http"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	contentTypeForm = "application/x-www-form-urlencoded"
)

// handleAuth saves a new AutoKitteh connection with user-submitted data.
func (h handler) handleSave(w http.ResponseWriter, r *http.Request) {
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

	apiKey := r.Form.Get("key_or_pat")
	vars := sdktypes.NewVars().Set(sdktypes.NewSymbol("APIKeyOrPAT"), apiKey, true)

	email := r.Form.Get("email")
	if email != "" {
		vars = vars.Set(sdktypes.NewSymbol("Email"), email, true)
	}

	sdkintegrations.FinalizeConnectionInit(w, r, integrationID, vars)
}
