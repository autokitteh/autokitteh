package http

import (
	"encoding/base64"
	"net/http"
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
	c.Finalize(sdktypes.NewVars().Set(authVar, auth, true))
}
