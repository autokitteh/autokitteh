package connection

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

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
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", vs.GetValue(privateClientIDVar))
	data.Set("client_secret", vs.GetValue(privateClientSecretVar))
	data.Set("scope", "https://graph.microsoft.com/.default")

	if data.Get("client_id") == "" || data.Get("client_secret") == "" {
		return nil, errors.New("missing required connection variables")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request for token failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request for token failed: %s (%s)", resp.Status, body)
	}

	var t oauth2.Token
	if err := json.Unmarshal(body, &t); err != nil {
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
		t := common.FreshOAuthToken(ctx, l, svc.OAuth, svc.Vars, desc, vs)
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
