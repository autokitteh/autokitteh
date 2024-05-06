package http

import (
	"encoding/base64"
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

var authVar = sdktypes.NewSymbol("auth")

// handleAuth saves a new autokitteh connection with user-submitted data.
func (h handler) handleAuth(w http.ResponseWriter, r *http.Request) {
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
	basic := r.Form.Get("basic_username") + ":" + r.Form.Get("basic_password")
	bearer := r.Form.Get("bearer_access_token")

	auth := ""
	switch {
	case basic != ":":
		auth = "Basic " + base64.StdEncoding.EncodeToString([]byte(basic))
	case bearer != "":
		auth = "Bearer " + bearer
	}

	initData := sdktypes.NewVars().Set(authVar, auth, true)

	sdkintegrations.FinalizeConnectionInit(w, r, IntegrationID, initData)
}
