package webhooks

import (
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	// AuthPath is the URL path for our webhook to save a new autokitteh
	// connection, after the user submits their Twilio secrets.
	AuthPath = "/twilio/save"

	HeaderContentType = "Content-Type"
	ContentTypeForm   = "application/x-www-form-urlencoded"
)

type Vars struct {
	AccountSID string
	Username   string `vars:"secret"`
	Password   string `vars:"secret"`
}

// HandleAuth saves a new autokitteh connection with user-submitted Twilio secrets.
func (h handler) HandleAuth(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("urlPath", r.URL.Path))

	// Check "Content-Type" header.
	if r.Header.Get(HeaderContentType) != ContentTypeForm {
		l.Error("Unexpected header value",
			zap.String("header", HeaderContentType),
			zap.String("got", r.Header.Get(HeaderContentType)),
			zap.String("want", ContentTypeForm),
		)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Read and parse POST request body.
	err := r.ParseForm()
	if err != nil {
		l.Error("Failed to parse inbound HTTP request",
			zap.Error(err),
		)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	accountSID := r.Form.Get("account_sid")
	username := accountSID
	password := r.Form.Get("auth_token")
	if password == "" {
		username = r.Form.Get("api_key")
		password = r.Form.Get("api_secret")
	}

	// TODO: Test the authentication details.

	initData := sdktypes.EncodeVars(Vars{
		AccountSID: accountSID,
		Username:   username,
		Password:   password,
	})

	sdkintegrations.FinalizeConnectionInit(w, r, h.integrationID, initData)
}
