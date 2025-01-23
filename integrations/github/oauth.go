package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations"
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
	vars   sdkservices.Vars
}

func NewHandler(l *zap.Logger, o sdkservices.OAuth, v sdkservices.Vars) handler {
	return handler{logger: l, oauth: o, vars: v}
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
		l.Warn("invalid data in OAuth redirect request", zap.Error(err))
		c.AbortBadRequest("invalid data in OAuth redirect request")
		return
	}

	// Parse and validate the results.
	setupAction := data.Params.Get("setup_action")
	l = l.With(zap.String("setup_action", setupAction))
	switch setupAction {
	case "":
		l.Warn("missing GitHub app setup action in OAuth redirect request")
		c.AbortBadRequest("missing GitHub app setup action")
		return
	case "request":
		l.Warn("GitHub app installation by non-org admin")
		c.AbortBadRequest("you need to be a GitHub org admin to approve this")
		return
	}

	installID := data.Params.Get("installation_id")
	l = l.With(zap.String("installation_id", installID))
	if installID == "" {
		l.Warn("missing GitHub app installation ID in OAuth redirect request")
		c.AbortBadRequest("missing GitHub app installation ID")
		return
	}

	iid, err := strconv.ParseInt(installID, 10, 64)
	if err != nil {
		l.Warn("invalid GitHub app installation ID in OAuth redirect request")
		c.AbortBadRequest("invalid GitHub app installation ID")
		return
	}

	// Get authoritative installation details.
	cid, err := sdktypes.StrictParseConnectionID(c.ConnectionID)
	if err != nil {
		l.Warn("invalid connection ID", zap.Error(err))
		c.AbortBadRequest("invalid connection ID")
		return
	}

	vs, err := h.vars.Get(r.Context(), sdktypes.NewVarScopeID(cid))
	if err != nil {
		l.Warn("failed to get GitHub app ID", zap.Error(err))
		c.AbortBadRequest("failed to get GitHub app ID")
		return
	}

	appID := os.Getenv("GITHUB_APP_ID")
	if vs.GetValueByString("client_secret") != "" {
		// Use custom GitHub App credentials instead of environment variables
		appID = vs.GetValueByString("app_id")
	}

	aid, err := strconv.ParseInt(appID, 10, 64)
	if err != nil {
		l.Warn("invalid GitHub app ID", zap.Error(err))
		c.AbortBadRequest("invalid GitHub app ID")
		return
	}

	gh, err := NewClientFromGitHubAppID(aid, vs.GetValue(vars.PrivateKey))
	if err != nil {
		l.Warn("failed to initialize GitHub app client", zap.Error(err))
		c.AbortBadRequest("failed to initialize GitHub app client")
		return
	}

	i, _, err := gh.Apps.GetInstallation(r.Context(), iid)
	if err != nil {
		l.Warn("failed to get GitHub app installation details", zap.Error(err))
		c.AbortBadRequest("failed to get GitHub app installation details")
		return
	}

	if i.ID == nil || *i.ID != iid {
		l.Warn("GitHub app installation details not found", zap.Any("installation", i))
		c.AbortBadRequest("GitHub app installation details not found")
		return
	}

	if i.Account == nil || i.Account.Login == nil {
		l.Warn("GitHub app installation details missing required fields")
		c.AbortBadRequest("GitHub app installation details missing required fields")
		return
	}

	name := *i.Account.Login

	events := fmt.Sprintf("%s", i.Events)
	events = events[1 : len(events)-1]

	perms, err := json.Marshal(i.Permissions)
	if err != nil {
		perms = []byte(err.Error())
	}
	ps := strings.ReplaceAll(string(perms[1:len(perms)-1]), `"`, "")
	ps = strings.ReplaceAll(ps, ",", " ")

	c.Finalize(sdktypes.NewVars().
		Set(vars.AppID, appID, false).
		Set(vars.AppName, *i.AppSlug, false).
		Set(vars.AuthType, integrations.OAuth, false).
		Set(vars.InstallID, installID, false).
		Set(vars.TargetID, strconv.FormatInt(*i.TargetID, 10), false).
		Set(vars.TargetName, name, false).
		Set(vars.TargetType, *i.TargetType, false).
		Set(vars.RepoSelection, *i.RepositorySelection, false).
		Set(vars.Permissions, ps, false).
		Set(vars.Events, events, false).
		Set(vars.UpdatedAt, i.UpdatedAt.Format(time.RFC3339), false).
		Set(vars.InstallKey(appID, installID), name, false))
}
