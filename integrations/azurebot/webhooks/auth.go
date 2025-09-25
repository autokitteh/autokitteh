package webhooks

import (
	"net/http"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	// AuthPath is the URL path for our webhook to save a new autokitteh
	// connection, after the user submits their Azure Bot secrets.
	AuthPath = "/azurebot/save"

	headerContentType = "Content-Type"
	contentTypeForm   = "application/x-www-form-urlencoded"
)

type Vars struct {
	AppID       string `var:"app_id"`
	AppPassword string `var:"app_password,secret"`
	TenantID    string `var:"tenant_id"`
}

// HandleAuth saves a new autokitteh connection with user-submitted Azure Bot secrets.
func (h handler) HandleAuth(w http.ResponseWriter, r *http.Request) {
	c, l := sdkintegrations.NewConnectionInit(h.logger, w, r, h.integration)

	contentType := r.Header.Get(headerContentType)
	if !strings.HasPrefix(contentType, contentTypeForm) {
		c.AbortBadRequest("unexpected content type")
		return
	}

	if err := r.ParseForm(); err != nil {
		l.Warn("Failed to parse incoming HTTP request", zap.Error(err))
		c.AbortBadRequest("form parsing error")
		return
	}

	c.Finalize(sdktypes.EncodeVars(Vars{
		AppID:       r.Form.Get("app_id"),
		AppPassword: r.Form.Get("app_password"),
		TenantID:    r.Form.Get("tenant_id"),
	}))
}
