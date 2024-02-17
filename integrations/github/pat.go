package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v54/github"
	"go.uber.org/zap"
)

const (
	// patPath is the URL path for our webhook to save a new autokitteh
	// PAT-based connection, after the user submits it via a web form.
	patPath = "/github/save_pat"

	HeaderContentType = "Content-Type"
	ContentTypeForm   = "application/x-www-form-urlencoded"
)

// HandlePAT saves a new autokitteh connection with a user-submitted token.
func (h handler) HandlePAT(w http.ResponseWriter, r *http.Request) {
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
	pat := r.Form.Get("pat")
	webhook := r.Form.Get("webhook")
	secret := r.Form.Get("secret")

	// Test the PAT's usability and get authoritative metadata details.
	ctx := context.Background()
	client := github.NewTokenClient(ctx, pat)
	user, resp, err := client.Users.Get(ctx, "")
	if err != nil {
		l.Warn("Unusable Personal Access Token",
			zap.Error(err),
		)
		e := "Unusable PAT error: " + err.Error()
		u := fmt.Sprintf("%serror.html?error=%s", uiPath, url.QueryEscape(e))
		http.Redirect(w, r, u, http.StatusFound)
		return
	}

	// Save a new connection, and return to the user an autokitteh connection token.
	connToken, err := h.createPATConnection(pat, webhook, secret, user, resp)
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

func (h handler) createPATConnection(pat, webhook, secret string, user *github.User, resp *github.Response) (string, error) {
	path := strings.Split(webhook, "/")
	if len(path) == 0 {
		return "", errors.New("unexpected webhook URL without '/'")
	}
	token, err := h.secrets.Create(context.Background(), h.scope,
		// Connection token --> Personal Access Token (to call API methods).
		map[string]string{
			"PAT":        pat,
			"secret":     secret,
			"login":      *user.Login,
			"type":       *user.Type,
			"expires":    strconv.FormatBool(!resp.TokenExpiration.IsZero()),
			"expiration": resp.TokenExpiration.Format(time.RFC3339),
		},
		// Webhook path suffix --> connection token & webhook secret (to dispatch API events).
		fmt.Sprintf("webhooks/%s", path[len(path)-1]),
	)
	if err != nil {
		return "", err
	}
	return token, nil
}
