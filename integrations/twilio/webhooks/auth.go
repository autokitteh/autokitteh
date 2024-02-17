package webhooks

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

const (
	// AuthPath is the URL path for our webhook to save a new autokitteh
	// connection, after the user submits their Twilio secrets.
	AuthPath = "/twilio/save"

	HeaderContentType = "Content-Type"
	ContentTypeForm   = "application/x-www-form-urlencoded"
)

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

	// Save a new connection, and return to the user an autokitteh connection token.
	connToken, err := h.createConnection(accountSID, username, password)
	if err != nil {
		l.Warn("Failed to save new connection secrets",
			zap.Error(err),
		)
		http.Error(w, "Internal Server Error: create connection", http.StatusInternalServerError)
		return
	}

	// Redirect the user to a success page: give them the connection token.
	l.Debug("Saved new autokitteh connection",
		zap.String("connectionToken", connToken),
	)
	url := "/twilio/connect/success.html?token=" + connToken
	http.Redirect(w, r, url, http.StatusFound)
}

// createConnection calls the Create method in SecretsService.
func (h handler) createConnection(accountSID, username, password string) (string, error) {
	token, err := h.secrets.Create(context.Background(), "twilio",
		// Connection token --> auth token / API key (to call API methods).
		// Auth token: username = account SID, password = auto token.
		// API key: username = API key, password = API secret.
		map[string]string{
			"accountSID": accountSID,
			"username":   username,
			"password":   password,
		},
		// Twilio account SIDs --> connection token(s) (to dispatch API events).
		fmt.Sprintf("accounts/%s", accountSID),
	)
	if err != nil {
		return "", err
	}
	return token, nil
}
