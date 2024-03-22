package slack

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/integrations/slack/api/auth"
	"go.autokitteh.dev/autokitteh/integrations/slack/api/bots"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

const (
	// uiPath is the URL root path of a simple web UI to interact with
	// users at the beginning and the end of their 3-legged OAuth v2
	// flow to install a Slack app.
	uiPath = "/slack/connect/"

	// oauthPath is the URL path for our handler to save
	// new OAuth-based connections.
	oauthPath = "/slack/oauth"
)

// handler is an autokitteh webhook which implements [http.Handler]
// to receive and dispatch asynchronous event notifications.
type handler struct {
	logger  *zap.Logger
	secrets sdkservices.Secrets
	scope   string
}

func NewHandler(l *zap.Logger, sec sdkservices.Secrets, scope string) http.Handler {
	return handler{logger: l, secrets: sec, scope: scope}
}

// ServeHTTP receives an inbound redirect request from autokitteh's OAuth
// management service. This request contains an OAuth token (if the OAuth
// flow was successful) and form parameters for debugging and validation
// (either way). If all is well, it saves a new autokitteh connection.
// Either way, it redirects the user to success or failure webpages.
func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("urlPath", r.URL.Path))

	// Handle errors (e.g. the user didn't authorize us) based on:
	// https://developers.google.com/identity/protocols/oauth2/web-server#handlingresponse
	e := r.FormValue("error")
	if e != "" {
		l.Warn("OAuth redirect request reported an error",
			zap.Error(errors.New(e)),
		)
		u := uiPath + "error.html?error=" + url.QueryEscape(e)
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
	ctx := extrazap.AttachLoggerToContext(l, r.Context())
	authTest, err := auth.TestWithToken(ctx, h.secrets, h.scope, oauthToken.AccessToken)
	if err != nil {
		e := "OAuth token test failed: " + err.Error()
		u := uiPath + "error.html?error=" + url.QueryEscape(e)
		http.Redirect(w, r, u, http.StatusFound)
		return
	}

	botInfo, err := bots.InfoWithToken(ctx, h.secrets, h.scope, oauthToken.AccessToken, authTest)
	if err != nil {
		e := "Bot info request failed: " + err.Error()
		u := uiPath + "error.html?error=" + url.QueryEscape(e)
		http.Redirect(w, r, u, http.StatusFound)
		return
	}

	// Save the OAuth token, and return to the user an autokitteh connection token.
	connToken, err := h.createOAuthConnection(ctx, authTest, botInfo, oauthToken)
	if err != nil {
		l.Warn("Failed to save new connection secrets", zap.Error(err))
		e := "Connection saving error: " + err.Error()
		u := uiPath + "error.html?error=" + url.QueryEscape(e)
		http.Redirect(w, r, u, http.StatusFound)
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

func (h handler) createOAuthConnection(ctx context.Context, authTest *auth.TestResponse, botInfo *bots.InfoResponse, oauthToken *oauth2.Token) (string, error) {
	token, err := h.secrets.Create(ctx, h.scope,
		// Connection token --> OAUth token (to call API methods).
		map[string]string{
			// Slack.
			"appID":        botInfo.Bot.AppID,
			"enterpriseID": authTest.EnterpriseID,
			"teamID":       authTest.TeamID,
			// OAuth token.
			"accessToken":  oauthToken.AccessToken,
			"tokenType":    oauthToken.TokenType,
			"refreshToken": oauthToken.RefreshToken,
			"expiry":       oauthToken.Expiry.Format(time.RFC3339),
		},
		// Slack app IDs --> connection token(s) (to dispatch API events).
		appSecretName(botInfo.Bot.AppID, authTest.EnterpriseID, authTest.TeamID),
	)
	if err != nil {
		return "", err
	}
	return token, nil
}

func appSecretName(appID, enterpriseID, teamID string) string {
	s := fmt.Sprintf("apps/%s/%s/%s", appID, enterpriseID, teamID)
	// Slack enterprise ID is allowed to be empty.
	return strings.ReplaceAll(s, "//", "/")
}
