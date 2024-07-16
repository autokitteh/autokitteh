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
	c, l := sdkintegrations.NewConnectionInit(h.logger, w, r, desc)

	// Check "Content-Type" header.
	contentType := r.Header.Get(headerContentType)
	if !strings.HasPrefix(contentType, contentTypeForm) {
		c.Abort("unexpected content type")
		return
	}

	// Read and parse POST request body.
	if err := r.ParseForm(); err != nil {
		l.Warn("Failed to parse incoming HTTP request", zap.Error(err))
		c.Abort("form parsing error")
		return
	}

	initData := sdktypes.NewVars().
		Set(baseURL, r.Form.Get("base_url"), false).
		Set(token, r.Form.Get("token"), true)

	addr := r.Form.Get("email")
	if addr != "" {
		initData = initData.Set(email, addr, true)
	}

	c.Finalize(initData)
}
