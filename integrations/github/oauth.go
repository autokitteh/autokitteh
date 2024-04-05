package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/google/go-github/v60/github"
	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

const (
	// uiPath is the URL root path of a simple web UI to interact with users at the
	// beginning and the end of their GitHub app installation and 3-legged OAuth v2 flow.
	uiPath = "/github/connect/"

	// oauthPath is the URL path for our handler to save new OAuth-based connections.
	oauthPath = "/github/oauth"
)

// handler is an autokitteh webhook which implements [http.Handler]
// to receive and dispatch asynchronous event notifications.
type handler struct {
	logger  *zap.Logger
	secrets sdkservices.Secrets
	oauth   sdkservices.OAuth
	scope   string
}

func NewHandler(l *zap.Logger, s sdkservices.Secrets, o sdkservices.OAuth, scope string) handler {
	return handler{logger: l, secrets: s, oauth: o, scope: scope}
}

// handleOAuth receives an inbound redirect request from autokitteh's OAuth
// management service. This request contains an OAuth token (if the OAuth
// flow was successful) and form parameters for debugging and validation
// (either way). If all is well, it saves a new autokitteh connection.
// Either way, it redirects the user to success or failure webpages.
func (h handler) handleOAuth(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("urlPath", r.URL.Path))

	// Handle errors (e.g. the user didn't authorize us) based on:
	// https://developers.google.com/identity/protocols/oauth2/web-server#handlingresponse
	e := r.FormValue("error")
	if e != "" {
		l.Warn("OAuth redirect request reported an error",
			zap.Error(errors.New(e)),
		)
		http.Redirect(w, r, uiPath+"error.html", http.StatusFound)
		return
	}

	// Parse and validate the results.
	v := r.FormValue("installation_id")
	if v == "" {
		l.Warn("OAuth redirect request without installation_id parameter")
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		l.Warn("OAuth redirect request with invalid installation_id parameter",
			zap.String("installationID", v),
			zap.String("setupAction", r.FormValue("setup_action")),
		)
		u := fmt.Sprintf("%serror.html?error=%s", uiPath, e)
		http.Redirect(w, r, u, http.StatusFound)
		return
	}
	// TODO: It should be simpler to extract the token from the request.
	t, err := time.Parse(time.RFC3339Nano, unescape(r, "ak_token_expiry"))
	if err != nil {
		l.Warn("OAuth redirect request with invalid token expiry timestamp",
			zap.String("installationID", v),
			zap.String("setupAction", r.FormValue("setup_action")),
			zap.String("timestamp", r.FormValue("ak_token_exiry")),
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

	// Test the OAuth token's usability and get authoritative installation details:
	// https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/generating-a-user-access-token-for-a-github-app
	// https://docs.github.com/en/rest/apps/installations#list-app-installations-accessible-to-the-user-access-token
	ctx := r.Context()
	src := h.tokenSource(ctx, oauthToken)
	gh := github.NewClient(oauth2.NewClient(ctx, src))
	is, _, err := gh.Apps.ListUserInstallations(ctx, &github.ListOptions{})
	if err != nil {
		l.Warn("OAuth user token source error",
			zap.Error(err),
		)
		http.Error(w, "Internal Server Error: token source", http.StatusInternalServerError)
		return
	}
	foundInstallation := false
	var i *github.Installation
	for _, i = range is {
		if *i.ID != id {
			continue
		}
		l.Debug("Verified new GitHub app installation",
			zap.Int64p("id", i.ID),
			zap.Stringp("repositorySelection", i.RepositorySelection),
			zap.Int64p("targetID", i.TargetID),
			zap.Stringp("targetName", i.Account.Login),
			zap.Stringp("targetType", i.TargetType),
		)
		foundInstallation = true
		break
	}
	if !foundInstallation {
		l.Warn("Installation details not found",
			zap.String("installationID", v),
		)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Return to the user an autokitteh connection token.
	connToken, err := h.createOAuthConnection(ctx, i)
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

func (h handler) createOAuthConnection(ctx context.Context, i *github.Installation) (string, error) {
	appID := strconv.FormatInt(*i.AppID, 10)
	installID := strconv.FormatInt(*i.ID, 10)

	token, err := h.secrets.Create(ctx, h.scope,
		// Connection token --> OAUth token (to call API methods).
		map[string]string{
			"appID":     appID,
			"installID": installID,
			"targetID":  strconv.FormatInt(*i.TargetID, 10),
			"login":     *i.Account.Login,
			"type":      *i.Account.Type,
		},
		// GitHub app IDs --> connection token(s) (to dispatch API events).
		fmt.Sprintf("apps/%s/%s", appID, installID),
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
