package scheduler

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

const (
	HeaderContentType = "Content-Type"
	ContentTypeForm   = "application/x-www-form-urlencoded"
)

// handler is an autokitteh webhook which implements [http.Handler]
// to save data from web form submissions as connections.
type handler struct {
	logger  *zap.Logger
	secrets sdkservices.Secrets
	scope   string
}

func NewHTTPHandler(l *zap.Logger, s sdkservices.Secrets, scope string) http.Handler {
	return handler{logger: l, secrets: s, scope: scope}
}

// ServeHTTP saves a new autokitteh connection with user-submitted data.
func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	schedule := r.Form.Get("schedule")
	timezone := r.Form.Get("timezone")
	memo := r.Form.Get("memo")

	// Test the validity of the schedule string.
	id, err := cronTable.AddFunc(schedule, func() {})
	if err != nil {
		e := "Invalid schedule error: " + err.Error()
		u := fmt.Sprintf("%serror.html?error=%s", uiPath, url.QueryEscape(e))
		http.Redirect(w, r, u, http.StatusFound)
		return
	}
	cronTable.Remove(id)

	// Save a new connection, and return to the user an autokitteh connection token.
	connToken, err := h.createConnection(schedule, timezone, memo)
	if err != nil {
		l.Warn("Failed to save new connection secrets",
			zap.Error(err),
		)
		e := "Connection saving error: " + err.Error()
		u := fmt.Sprintf("%serror.html?error=%s", uiPath, url.QueryEscape(e))
		http.Redirect(w, r, u, http.StatusFound)
		return
	}

	// Redirect the user to a success page: give them the connection token.
	l.Debug("Saved new autokitteh connection")
	u := fmt.Sprintf("%ssuccess.html?token=%s", uiPath, connToken)
	http.Redirect(w, r, u, http.StatusFound)
}

func (h handler) createConnection(schedule, timezone, memo string) (string, error) {
	token, err := h.secrets.Create(context.Background(), h.scope,
		// Connection token --> schedule details.
		map[string]string{
			"schedule": schedule,
			"timezone": timezone,
			"memo":     memo,
		},
		// List of all connection tokens.
		"tokens",
	)
	if err != nil {
		return "", err
	}
	return token, nil
}
