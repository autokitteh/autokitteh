package http

import (
	"encoding/base64"
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
)

var authVar = sdktypes.NewSymbol("auth")

// handleAuth saves a new autokitteh connection with user-submitted data.
func (h handler) handleAuth(w http.ResponseWriter, r *http.Request) {
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
		msg := url.QueryEscape(err.Error())
		u := fmt.Sprintf("%s/error.html?error=%s", desc.ConnectionURL().Path, msg)
		http.Redirect(w, r, u, http.StatusFound)
		return
	}

	basic := r.Form.Get("basic_username") + ":" + r.Form.Get("basic_password")
	bearer := r.Form.Get("bearer_access_token")

	auth := ""
	switch {
	case basic != ":":
		auth = "Basic " + base64.StdEncoding.EncodeToString([]byte(basic))
	case bearer != "":
		auth = "Bearer " + bearer
	}

	// Finalize the connection initialization - save the auth data in the connection.
	initData := sdktypes.NewVars().Set(authVar, auth, true)

	sdkintegrations.FinalizeConnectionInit(w, r, IntegrationID, initData)
}
