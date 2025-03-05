package salesforce

import (
	"context"
	"time"

	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// oauthToken returns the OAuth token stored in the
// connection variables. If it's stale, we refresh it first.
func oauthToken(ctx context.Context, vs sdktypes.Vars, o sdkservices.OAuth) *oauth2.Token {
	t1 := &oauth2.Token{
		AccessToken:  vs.GetValue(common.OAuthAccessTokenVar),
		RefreshToken: vs.GetValue(common.OAuthRefreshTokenVar),
		TokenType:    vs.GetValue(common.OAuthTokenTypeVar),
	}

	if expiryStr := vs.GetValue(common.OAuthExpiryVar); expiryStr != "" {
		if expiry, err := time.Parse(time.RFC3339, expiryStr); err == nil {
			t1.Expiry = expiry
		} else {
			// TODO: Question: Should we force a refresh here?
			t1.Expiry = time.Now().Add(time.Hour * 2)
		}
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

	switch authType := common.ReadAuthType(vs); authType {
	case integrations.OAuthPrivate:
		return "Bearer " + oauthToken(ctx, vs, h.oauth).AccessToken
	// Unknown/unrecognized mode - an error.
	default:
		l.Error("Salesforce subscription: unexpected auth type", zap.String("auth_type", authType))
		return ""
	}
}
