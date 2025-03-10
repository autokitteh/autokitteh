package microsoft

import (
	"context"
	"errors"
	"net/http"

	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/microsoft/connection"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// handleOAuth receives an incoming redirect request from AutoKitteh's
// generic OAuth service, which contains an OAuth token (if the OAuth
// flow was successful) and form parameters for debugging and validation.
// This is the last step in a 3-legged OAuth 2.0 flow, in which we verify
// the usability of the OAuth token, and save it as connection variables.
func (h handler) handleOAuth(w http.ResponseWriter, r *http.Request) {
	c, l := sdkintegrations.NewConnectionInit(h.logger, w, r, desc)

	// Parse the GET request's query params.
	if err := r.ParseForm(); err != nil {
		l.Warn("save connection after OAuth flow: failed to parse HTTP request", zap.Error(err))
		c.AbortBadRequest("request parsing error")
		return
	}

	// Sanity check: the connection ID is valid.
	cid, err := sdktypes.StrictParseConnectionID(c.ConnectionID)
	if err != nil {
		l.Warn("save connection after OAuth flow: invalid connection ID", zap.Error(err))
		c.AbortBadRequest("invalid connection ID")
		return
	}

	// Handle OAuth errors (e.g. the user didn't authorize us), based on:
	// https://developers.google.com/identity/protocols/oauth2/web-server#handlingresponse
	e := r.FormValue("error")
	if e != "" {
		l.Warn("OAuth redirection reported an error", zap.String("error", e))
		c.AbortBadRequest(e)
		return
	}

	// Decode the OAuth token.
	var data sdkintegrations.OAuthData
	err = kittehs.DecodeURLData(r.FormValue("oauth"), &data)
	if err != nil {
		l.Error("OAuth redirection returned invalid results", zap.Error(err))
		c.AbortServerError("invalid OAuth data")
		return
	}

	// Test the token's usability and get authoritative connection details.
	ctx := r.Context()
	org, err := connection.GetOrgInfo(ctx, data.Token)
	if err != nil {
		l.Warn("MS organization details request failed", zap.Error(err))
		c.AbortServerError("organization details request failed")
		return
	}

	user, err := connection.GetUserInfo(ctx, data.Token)
	if err != nil {
		l.Warn("MS user details request failed", zap.Error(err))
		c.AbortServerError("user details request failed")
		return
	}

	vsid := sdktypes.NewVarScopeID(cid)
	if err := h.saveConnection(ctx, vsid, data.Token, org, user); err != nil {
		l.Error("failed to save OAuth connection details", zap.Error(err))
		c.AbortServerError("failed to save connection details")
		return
	}

	// Subscribe to receive asynchronous change notifications from
	// Microsoft Graph, based on the connection's integration type.
	svc := connection.NewServices(l, h.vars, h.oauth)
	_ = errors.Join(connection.Subscribe(ctx, svc, cid, resources(c.Integration))...)
	// TODO(INT-232): "Subscription operations for tenant-wide chats subscription is not allowed in 'OnBehalfOfUser' context."
	// if err != nil {
	// 	c.AbortServerError("failed to create/renew event subscriptions")
	// 	return
	// }

	// Redirect the user back to the UI.
	urlPath, err := c.FinalURL()
	if err != nil {
		l.Error("failed to construct final OAuth URL", zap.Error(err))
		c.AbortServerError("bad redirect URL")
		return
	}

	http.Redirect(w, r, urlPath, http.StatusFound)
}

// saveConnection saves OAuth token details as connection variables.
func (h handler) saveConnection(ctx context.Context, vsid sdktypes.VarScopeID, t *oauth2.Token, o *connection.OrgInfo, u *connection.UserInfo) error {
	if t == nil {
		return errors.New("OAuth redirection missing token data")
	}

	vs := sdktypes.EncodeVars(common.EncodeOAuthData(t))
	vs = vs.Append(sdktypes.EncodeVars(o)...)
	vs = vs.Append(sdktypes.EncodeVars(u)...)
	return h.vars.Set(ctx, vs.WithScopeID(vsid)...)
}
