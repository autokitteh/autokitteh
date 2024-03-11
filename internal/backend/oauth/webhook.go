package oauth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

const (
	exchangeTimeout = 3 * time.Second
)

type svc struct {
	logger *zap.Logger
	svcs   sdkservices.Services
}

func NewWebhook(l *zap.Logger, c sdkservices.Services) http.Handler {
	return &svc{logger: l, svcs: c}
}

func (s *svc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Parse the request's URL path (skip the "/oauth" prefix).
	path := strings.TrimPrefix(r.URL.Path, "/oauth/")
	action, id, found := strings.Cut(path, "/")
	if !found || action == "" || id == "" {
		s.logger.Warn("Incomplete path in OAuth request URL",
			zap.String("urlPath", r.URL.Path),
		)
		http.Error(w, "Bad request: incomplete URL path", http.StatusBadRequest)
		return
	}

	// Sanity check: the specified OAuth consumer is registered.
	o := s.svcs.OAuth()
	ctx := context.Background()
	_, _, err := o.Get(ctx, id)
	if err != nil {
		s.logger.Warn("OAuth config retrieval error",
			zap.String("urlPath", r.URL.Path),
			zap.Error(err),
		)
		http.Error(w, "Bad request: unrecognized config ID", http.StatusBadRequest)
		return
	}

	switch action {
	// Start a 3-legged OAuth v2 user flow.
	case "start":
		s.startFlow(ctx, w, r, o, id)
	// Accept a temporary authorization code via redirection,
	// and convert it into a refresh token / user access token.
	case "redirect":
		s.exchangeCode(ctx, w, r, o, id)
	// Error: bad request with an unrecognized action.
	default:
		s.logger.Warn("Unrecognized OAuth webhook action",
			zap.String("urlPath", r.URL.Path),
		)
		http.Error(w, "Bad request: unrecognized webhook action", http.StatusBadRequest)
	}
}

func (s *svc) startFlow(ctx context.Context, w http.ResponseWriter, r *http.Request, o sdkservices.OAuth, id string) {
	startURL, err := o.StartFlow(ctx, id)
	if err != nil {
		s.logger.Error("OAuth start flow error",
			zap.String("urlPath", r.URL.Path),
			zap.Error(err),
		)
		http.Error(w, "Bad request: start flow error", http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, startURL, http.StatusFound)
}

func (s *svc) exchangeCode(ctx context.Context, w http.ResponseWriter, r *http.Request, o sdkservices.OAuth, id string) {
	destURL := fmt.Sprintf("/%s/oauth?%s", id, r.URL.RawQuery)
	destURL = regexp.MustCompile(`&?code=[^&]*`).ReplaceAllString(destURL, "")
	destURL = regexp.MustCompile(`&?state=[^&]*`).ReplaceAllString(destURL, "")

	// Read our state value from the the redirected request.
	state := r.FormValue("state")
	if state == "" {
		s.logger.Warn("OAuth redirect request without state parameter",
			zap.String("urlPath", r.URL.Path),
			zap.Any("form", r.Form),
		)
		destURL += "&error=" + url.QueryEscape("Missing state parameter")
		http.Redirect(w, r, destURL, http.StatusFound)
		return
	}

	// Read the temporary authorization code from the redirected request.
	code := r.FormValue("code")
	if code == "" {
		s.logger.Warn("OAuth redirect request without code parameter",
			zap.String("urlPath", r.URL.Path),
			zap.Any("form", r.Form),
		)
		destURL += "&error=" + url.QueryEscape("Missing code parameter")
		http.Redirect(w, r, destURL, http.StatusFound)
		return
	}

	// Convert it into an OAuth refresh token / user access token.
	token, err := o.Exchange(ctx, id, state, code)
	if err != nil {
		s.logger.Warn("OAuth exchange error",
			zap.String("urlPath", r.URL.Path),
			zap.Error(err),
		)
		destURL += "&error=" + url.QueryEscape(err.Error())
		http.Redirect(w, r, destURL, http.StatusFound)
		return
	}

	// Pass the new OAuth token along with the redirected request's
	// URL parameters to the initiator of the OAuth flow.
	destURL += "&ak_token_access=" + url.QueryEscape(token.AccessToken)
	destURL += "&ak_token_refresh=" + url.QueryEscape(token.RefreshToken)
	destURL += "&ak_token_type=" + url.QueryEscape(token.TokenType)
	destURL += "&ak_token_expiry=" + url.QueryEscape(token.Expiry.UTC().Format(time.RFC3339Nano))
	destURL = strings.ReplaceAll(destURL, "?&", "?")
	http.Redirect(w, r, destURL, http.StatusFound)
}
