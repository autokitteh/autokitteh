package salesforce

import (
	"context"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

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

	t := common.FreshOAuthToken(ctx, l, h.oauth, h.vars, desc, vs)
	return "Bearer " + t.AccessToken
}
