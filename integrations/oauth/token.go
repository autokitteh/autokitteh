package oauth

import (
	"context"
	"time"

	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// FreshToken returns the OAuth token stored in the connection variables.
// If it's stale, this function refreshes it first. A token without an expiry
// is considered fresh forever, so time-limited tokens with a missing timestamp
// need to add it. Refreshed tokens are saved back to the connection variables.
func (o *OAuth) FreshToken(ctx context.Context, l *zap.Logger, i sdktypes.Integration, vs sdktypes.Vars) *oauth2.Token {
	data := new(common.OAuthData)
	vs.Decode(data)
	t1 := data.ToToken()

	// Access token is still fresh - return it as-is.
	if t1.Valid() {
		return t1
	}

	// Otherwise, use the OAuth refresh flow.
	integ := i.UniqueName().String()
	vsid := vs.Get(common.OAuthAccessTokenVar).ScopeID()
	cid := vsid.ToConnectionID()
	l = l.With(
		zap.String("integration", integ),
		zap.String("connection_id", vsid.String()),
	)

	cfg, _, err := o.GetConfig(ctx, integ, cid)
	if err != nil {
		l.Error("failed to get OAuth config to refresh a token", zap.Error(err))
		return t1
	}

	t2, err := cfg.TokenSource(ctx, t1).Token()
	if err != nil {
		l.Warn("failed to refresh OAuth token, returning original one", zap.Error(err))
		return t1
	}

	// Special case: Salesforce access tokens are time-limited and yet
	// they don't have an expiry timestamp - so we add it on our own.
	if o.flags(integ).expiryMissingInToken && t2.Expiry.IsZero() {
		// TODO(INT-322): Reuse "accessTokenExpiration" in SFDC's OAuth handler.
		t2.Expiry = time.Now().UTC().Add(2 * time.Hour)
	}

	l.Debug("refreshed OAuth token", zap.Time("new_expiry", t2.Expiry))

	// Update the connection variables before returning the new token.
	ctx = authcontext.SetAuthnSystemUser(ctx)
	vs = sdktypes.EncodeVars(common.EncodeOAuthData(t2))
	if err := o.vars.Set(ctx, vs.WithScopeID(vsid)...); err != nil {
		l.Error("failed to save refreshed OAuth token in connection", zap.Error(err))
	}

	return t2
}
