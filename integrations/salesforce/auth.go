package salesforce

import (
	"context"
	"time"

	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// oauthToken returns the OAuth token stored in the
// connection variables. If it's stale, we refresh it first.
func oauthToken(ctx context.Context, vs sdktypes.Vars, o sdkservices.OAuth) *oauth2.Token {
	exp, err := time.Parse(time.RFC3339, vs.GetValue(common.OAuthExpiryVar))
	if err != nil {
		exp = time.Now().UTC().Add(-time.Minute)
	}

	t1 := &oauth2.Token{
		AccessToken:  vs.GetValue(common.OAuthAccessTokenVar),
		RefreshToken: vs.GetValue(common.OAuthRefreshTokenVar),
		TokenType:    vs.GetValue(common.OAuthTokenTypeVar),
		Expiry:       exp,
	}

	if t1.Valid() {
		return t1
	}

	cfg, _, err := o.Get(ctx, "salesforce")
	if err != nil {
		return t1
	}

	t2, err := cfg.TokenSource(ctx, t1).Token()
	if err != nil {
		return t1
	}

	return t2
}

// bearerToken returns a bearer token for an HTTP Authorization header based on the connection's auth type.
func (h handler) bearerToken(ctx context.Context, l *zap.Logger, cid sdktypes.ConnectionID) string {
	ctx = authcontext.SetAuthnSystemUser(ctx)

	vs, err := h.vars.Get(ctx, sdktypes.NewVarScopeID(cid))
	if err != nil {
		l.Error("failed to read connection vars",
			zap.String("connection_id", cid.String()), zap.Error(err),
		)
		return ""
	}

	return "Bearer " + oauthToken(ctx, vs, h.oauth).AccessToken
}
