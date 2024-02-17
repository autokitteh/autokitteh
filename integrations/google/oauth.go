package google

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
	googleoauth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

const (
	// uiPath is the URL root path of a simple web UI to interact with users at
	// the beginning and the end of their 3-legged OAuth v2 flow with Google.
	uiPath = "/google/connect/"

	// oauthPath is the URL path for our handler to save new OAuth-based connections.
	oauthPath = "/google/oauth"
)

// handler is an autokitteh webhook which implements [http.Handler]
// to receive and dispatch asynchronous event notifications.
type handler struct {
	logger  *zap.Logger
	secrets sdkservices.Secrets
	oauth   sdkservices.OAuth
	scope   string
}

func NewHTTPHandler(l *zap.Logger, sec sdkservices.Secrets, o sdkservices.OAuth, scope string) handler {
	return handler{logger: l, secrets: sec, oauth: o, scope: scope}
}

// HandleOAuth receives an inbound redirect request from autokitteh's OAuth
// management service. This request contains an OAuth token (if the OAuth
// flow was successful) and form parameters for debugging and validation
// (either way). If all is well, it saves a new autokitteh connection.
// Either way, it redirects the user to success or failure webpages.
func (h handler) HandleOAuth(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("urlPath", r.URL.Path))

	// Handle errors (e.g. the user didn't authorize us) based on:
	// https://developers.google.com/identity/protocols/oauth2/web-server#handlingresponse
	e := r.FormValue("error")
	if e != "" {
		l.Warn("OAuth redirect request reported an error",
			zap.Error(errors.New(e)),
		)
		u := fmt.Sprintf("%serror.html?error=%s", uiPath, e)
		http.Redirect(w, r, u, http.StatusFound)
		return
	}

	// Parse and validate the results.
	// TODO: It should be simpler to extract the token from the request.
	t, err := time.Parse(time.RFC3339Nano, unescape(r, "ak_token_expiry"))
	if err != nil {
		l.Warn("OAuth redirect request with invalid token expiry timestamp",
			zap.String("timestamp", r.FormValue("ak_token_expiry")),
		)
		u := uiPath + "error.html?error=" + url.QueryEscape("invalid OAuth token timestamp")
		http.Redirect(w, r, u, http.StatusFound)
		return
	}
	oauthToken := &oauth2.Token{
		AccessToken:  unescape(r, "ak_token_access"),
		RefreshToken: unescape(r, "ak_token_refresh"),
		TokenType:    unescape(r, "ak_token_type"),
		Expiry:       t,
	}

	// Test the OAuth token's usability and get authoritative installation details.
	src := h.tokenSource(r.Context(), oauthToken)
	svc, err := googleoauth2.NewService(r.Context(), option.WithTokenSource(src))
	if err != nil {
		l.Warn("OAuth user token error",
			zap.Error(err),
		)
		http.Error(w, "Internal Server Error: token source", http.StatusInternalServerError)
		return
	}
	ui, ti, err := h.getUserDetails(r.Context(), w, svc)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Save the OAuth token, and return to the user an autokitteh connection token.
	connToken, err := h.createConnection(r.Context(), ui, ti, oauthToken)
	if err != nil {
		l.Warn("Failed to save new connection secrets",
			zap.Error(err),
		)
		http.Error(w, "Internal Server Error: create connection", http.StatusInternalServerError)
		return
	}

	// Redirect the user to a success page: give them the connection token.
	l.Debug("Completed OAuth flow")
	u := fmt.Sprintf("%ssuccess.html?token=%s", uiPath, connToken)
	http.Redirect(w, r, u, http.StatusFound)
}

// unescape returns a named URL-unescaped query parameter value,
// or an empty string if it's missing, or URL-escaped improperly.
func unescape(r *http.Request, key string) string {
	s, err := url.QueryUnescape(r.FormValue(key))
	if err != nil {
		return ""
	}
	return s
}

func (h handler) createConnection(ctx context.Context, u *googleoauth2.Userinfo, t *googleoauth2.Tokeninfo, oauthToken *oauth2.Token) (string, error) {
	token, err := h.secrets.Create(ctx, h.scope,
		// Connection token --> OAUth token (to call API methods).
		map[string]string{
			// Google.
			"userID":     u.Id,
			"email":      u.Email,
			"name":       u.Name,
			"givenName":  u.GivenName,
			"familyName": u.FamilyName,
			"hd":         u.Hd,
			"scopes":     t.Scope,
			// OAuth token.
			"accessToken":  oauthToken.AccessToken,
			"tokenType":    oauthToken.TokenType,
			"refreshToken": oauthToken.RefreshToken,
			"expiry":       oauthToken.Expiry.Format(time.RFC3339),
		},
		// Google user ID --> connection token(s) (to dispatch API events).
		fmt.Sprintf("users/%s", u.Id),
	)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (h handler) tokenSource(ctx context.Context, t *oauth2.Token) oauth2.TokenSource {
	cfg, _, err := h.oauth.Get(ctx, h.scope)
	if err != nil {
		return nil
	}
	return cfg.TokenSource(ctx, t)
}
