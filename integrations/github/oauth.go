package github

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/google/go-github/v60/github"
	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"go.autokitteh.dev/autokitteh/integrations/github/internal/vars"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
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
	c, l := sdkintegrations.NewConnectionInit(h.logger, w, r, desc)

	// Handle errors (e.g. the user didn't authorize us) based on:
	// https://developers.google.com/identity/protocols/oauth2/web-server#handlingresponse
	e := r.FormValue("error")
	if e != "" {
		l.Warn("OAuth redirect request reported an error", zap.Error(errors.New(e)))
		c.AbortBadRequest(e)
		return
	}

	_, data, err := sdkintegrations.GetOAuthDataFromURL(r.URL)
	if err != nil {
		l.Warn("Invalid data in OAuth redirect request", zap.Error(err))
		c.AbortBadRequest("invalid data parameter")
		return
	}

	// Parse and validate the results.
	v := data.Params.Get("installation_id")
	if v == "" {
		l.Warn("Missing installation ID in OAuth redirect request")
		c.AbortBadRequest("missing GitHub installation ID")
		return
	}

	l = l.With(zap.String("installation_id", v))

	id, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		l.Warn("Invalid GitHub installation ID in OAuth redirect request",
			zap.String("setupAction", r.FormValue("setup_action")),
		)
		c.AbortBadRequest("invalid GitHub installation ID")
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
		l.Warn("GitHub enterprise URL error", zap.Error(err))
		c.AbortBadRequest("token source")
		return
	}
	if u != "" {
		gh, err = gh.WithEnterpriseURLs(u, u)
		if err != nil {
			l.Warn("GitHub enterprise URL error", zap.String("url", u), zap.Error(err))
			c.AbortBadRequest(err.Error())
			return
		}
	}

	// TODO: Support pagination.
	is, _, err := gh.Apps.ListUserInstallations(ctx, &github.ListOptions{})
	if err != nil {
		l.Warn("OAuth user token source error", zap.Error(err))
		c.AbortBadRequest("list installations error")
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
		l.Warn("GitHub app installation details not found")
		c.AbortBadRequest("GitHub app installation details not found")
		return
	}

	if i.AppID == nil || i.ID == nil || i.Account == nil || i.Account.Login == nil {
		l.Warn("GitHub app installation details missing required fields")
		c.AbortBadRequest("GitHub app installation details missing required fields")
		return
	}

	appID := strconv.FormatInt(*i.AppID, 10)
	installID := strconv.FormatInt(*i.ID, 10)
	user := string(*i.Account.Login)

	c.Finalize(sdktypes.NewVars().
		Set(vars.UserAppID(user), appID, false).
		Set(vars.UserInstallID(user), installID, false).
		Set(vars.InstallKey(appID, installID), user, false))
}

func (h handler) tokenSource(ctx context.Context, t *oauth2.Token) oauth2.TokenSource {
	cfg, _, err := h.oauth.Get(ctx, "github")
	if err != nil {
		return nil
	}
	return cfg.TokenSource(ctx, t)
}
