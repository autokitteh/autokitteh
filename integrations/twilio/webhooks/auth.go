package webhooks

import (
	"net/http"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var AuthType = sdktypes.NewSymbol("authType")

const (
	// AuthPath is the URL path for our webhook to save a new autokitteh
	// connection, after the user submits their Twilio secrets.
	AuthPath = "/twilio/save"

	headerContentType = "Content-Type"
	contentTypeForm   = "application/x-www-form-urlencoded"
)

type Vars struct {
	AccountSID string
	Username   string `vars:"secret"`
	Password   string `vars:"secret"`
}

// HandleAuth saves a new autokitteh connection with user-submitted Twilio secrets.
func (h handler) HandleAuth(w http.ResponseWriter, r *http.Request) {
	c, l := sdkintegrations.NewConnectionInit(h.logger, w, r, h.integration)

	// Check "Content-Type" header.
	contentType := r.Header.Get(headerContentType)
	if !strings.HasPrefix(contentType, contentTypeForm) {
		c.AbortBadRequest("unexpected content type")
		return
	}

	// Read and parse POST request body.
	if err := r.ParseForm(); err != nil {
		l.Warn("Failed to parse incoming HTTP request", zap.Error(err))
		c.AbortBadRequest("form parsing error")
		return
	}

	at := ""

	accountSID := r.Form.Get("account_sid")
	username := accountSID
	password := r.Form.Get("auth_token")
	if password == "" {
		username = r.Form.Get("api_key")
		password = r.Form.Get("api_secret")
		at = integrations.APIKey
	} else {
		at = integrations.APIToken
	}

	// TODO(ENG-1156): Test the authentication details.

	c.Finalize(sdktypes.EncodeVars(Vars{
		AccountSID: accountSID,
		Username:   username,
		Password:   password,
	}).Set(AuthType, at, false))
}
