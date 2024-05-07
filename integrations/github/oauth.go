package github

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/google/go-github/v60/github"
	"go.uber.org/zap"

	"golang.org/x/oauth2"

	"go.autokitteh.dev/autokitteh/integrations/github/internal/vars"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
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
	logger *zap.Logger
	oauth  sdkservices.OAuth
}

func NewHandler(l *zap.Logger, o sdkservices.OAuth) handler {
	return handler{logger: l, oauth: o}
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
		l.Warn("OAuth redirect request reported an error", zap.Error(errors.New(e)))
		u := uiPath + "error.html?error=" + url.QueryEscape(e)
		http.Redirect(w, r, u, http.StatusFound)
		return
	}

	_, data, err := sdkintegrations.GetOAuthDataFromURL(r.URL)
	if err != nil {
		l.Warn("OAuth redirect request with invalid data parameter", zap.Error(err))
		u := uiPath + "error.html?error=" + url.QueryEscape("invalid data parameter")
		http.Redirect(w, r, u, http.StatusFound)
		return
	}

	// Parse and validate the results.
	v := data.Params.Get("installation_id")
	if v == "" {
		l.Warn("OAuth redirect request without installation_id parameter")
		u := uiPath + "error.html?error=" + url.QueryEscape("missing installation ID")
		http.Redirect(w, r, u, http.StatusFound)
		return
	}

	l = l.With(zap.String("installation_id", v))

	id, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		l.Warn("OAuth redirect request with invalid installation_id parameter",
			zap.String("setupAction", r.FormValue("setup_action")),
		)
		u := uiPath + "error.html?error=" + url.QueryEscape("invalid installation ID")
		http.Redirect(w, r, u, http.StatusFound)
		return
	}

	// Test the OAuth token's usability and get authoritative installation details:
	// https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/generating-a-user-access-token-for-a-github-app
	// https://docs.github.com/en/rest/apps/installations#list-app-installations-accessible-to-the-user-access-token
	ctx := r.Context()
	src := h.tokenSource(ctx, data.Token)
	gh := github.NewClient(oauth2.NewClient(ctx, src))
	u, err := enterpriseURL()
	if err != nil {
		l.Warn("Enterprise URL error", zap.Error(err))
		u := uiPath + "error.html?error=" + url.QueryEscape(err.Error())
		http.Redirect(w, r, u, http.StatusFound)
		return
	}
	if u != "" {
		gh, err = gh.WithEnterpriseURLs(u, u)
		if err != nil {
			l.Warn("Enterprise URL error", zap.String("url", u), zap.Error(err))
			u := uiPath + "error.html?error=" + url.QueryEscape(err.Error())
			http.Redirect(w, r, u, http.StatusFound)
			return
		}
	}

	// TODO: Go over all pages.
	is, _, err := gh.Apps.ListUserInstallations(ctx, &github.ListOptions{})
	if err != nil {
		l.Warn("OAuth user token source error", zap.Error(err))
		u := uiPath + "error.html?error=" + url.QueryEscape(err.Error())
		http.Redirect(w, r, u, http.StatusFound)
		return
	}
	foundInstallation := false
	var i *github.Installation
	for _, i = range is {
		if i.ID == nil || *i.ID != id {
			continue
		}
		l.Debug("Verified new GitHub app installation",
			zap.Stringp("repositorySelection", i.RepositorySelection),
			zap.Int64p("targetID", i.TargetID),
			zap.Stringp("targetName", i.Account.Login),
			zap.Stringp("targetType", i.TargetType),
		)
		foundInstallation = true
		break
	}
	if !foundInstallation {
		l.Warn("Installation details not found")
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if i.AppID == nil || i.ID == nil || i.Account == nil || i.Account.Login == nil {
		l.Warn("Installation details missing required fields")
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	appID := strconv.FormatInt(*i.AppID, 10)
	installID := strconv.FormatInt(*i.ID, 10)
	user := string(*i.Account.Login)

	initData := sdktypes.NewVars().
		Set(vars.UserAppID(user), appID, false).
		Set(vars.UserInstallID(user), installID, false).
		Set(vars.InstallKey(appID, installID), user, false)

	sdkintegrations.FinalizeConnectionInit(w, r, integrationID, initData)
}

func (h handler) tokenSource(ctx context.Context, t *oauth2.Token) oauth2.TokenSource {
	cfg, _, err := h.oauth.Get(ctx, "github")
	if err != nil {
		return nil
	}
	return cfg.TokenSource(ctx, t)
}
