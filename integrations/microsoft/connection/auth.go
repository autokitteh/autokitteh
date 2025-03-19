package connection

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// DaemonToken returns a Microsoft Graph daemon app token, which is the
// same as a regular OAuth token, but without the 3-legged OAuth 2.0 flow
// (based on: https://learn.microsoft.com/en-us/entra/identity-platform/scenario-daemon-acquire-token).
func DaemonToken(ctx context.Context, vs sdktypes.Vars) (*oauth2.Token, error) {
	// https://learn.microsoft.com/en-us/answers/questions/1853467/aadsts7000229-the-client-application-is-missing-se
	tenantID := vs.GetValue(privateTenantIDVar)
	if tenantID == "" {
		tenantID = vs.GetValue(orgIDVar)
	}
	if tenantID == "" {
		return nil, errors.New("missing tenant ID")
	}

	u := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tenantID)

	// TODO(INT-227): Add support for certificate-based authentication.
	form := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {vs.GetValue(common.PrivateClientIDVar)},
		"client_secret": {vs.GetValue(common.PrivateClientSecretVar)},
		"scope":         {"https://graph.microsoft.com/.default"},
	}

	if form.Get("client_id") == "" || form.Get("client_secret") == "" {
		return nil, errors.New("missing required connection variables")
	}

	resp, err := common.HTTPPostForm(ctx, u, "", form)
	if err != nil {
		return nil, err
	}

	var t oauth2.Token
	if err := json.Unmarshal(resp, &t); err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	return &t, nil
}

// bearerToken returns a bearer token for an HTTP Authorization header based on the connection's auth type.
func bearerToken(ctx context.Context, l *zap.Logger, svc Services, cid sdktypes.ConnectionID) string {
	ctx = authcontext.SetAuthnSystemUser(ctx)

	vs, err := svc.Vars.Get(ctx, sdktypes.NewVarScopeID(cid))
	if err != nil {
		l.Error("failed to read connection vars",
			zap.String("connection_id", cid.String()), zap.Error(err),
		)
		return ""
	}

	switch authType := common.ReadAuthType(vs); authType {
	case integrations.OAuthDefault, integrations.OAuthPrivate:
		desc := common.Descriptor("microsoft", "", "")
		t := svc.OAuth.FreshToken(ctx, l, desc, vs)
		return "Bearer " + t.AccessToken

	case integrations.DaemonApp:
		t, err := DaemonToken(ctx, vs)
		if err != nil {
			return ""
		}
		return "Bearer " + t.AccessToken

	// Unknown/unrecognized mode - an error.
	default:
		l.Error("MS Graph subscription: unexpected auth type", zap.String("auth_type", authType))
		return ""
	}
}
