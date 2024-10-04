package gmail

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	googleoauth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"

	"go.autokitteh.dev/autokitteh/integrations/google/connections"
	"go.autokitteh.dev/autokitteh/integrations/google/internal/vars"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type api struct {
	vars sdkservices.Vars
	cid  sdktypes.ConnectionID
}

var IntegrationID = sdktypes.NewIntegrationIDFromName("gmail")

var desc = kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
	IntegrationId: IntegrationID.String(),
	UniqueName:    "gmail",
	DisplayName:   "Gmail",
	Description:   "Gmail is an email service provided by Google.",
	LogoUrl:       "/static/images/gmail.svg",
	ConnectionUrl: "/gmail/connect",
	ConnectionCapabilities: &sdktypes.ConnectionCapabilitiesPB{
		RequiresConnectionInit: true,
	},
}))

func New(cvars sdkservices.Vars) sdkservices.Integration {
	return sdkintegrations.NewIntegration(
		desc,
		sdkmodule.Empty,
		connections.ConnStatus(cvars),
		connections.ConnTest(cvars),
		sdkintegrations.WithConnectionConfigFromVars(cvars))
}

func (a api) gmailClient(ctx context.Context) (*gmail.Service, error) {
	data, err := a.connectionData(ctx)
	if err != nil {
		return nil, err
	}

	var src oauth2.TokenSource
	if data.OAuthData != "" {
		if src, err = a.oauthTokenSource(ctx, data.OAuthData); err != nil {
			return nil, err
		}
	} else {
		src, err = a.jwtTokenSource(ctx, data.JSON)
		if err != nil {
			return nil, err
		}
	}

	svc, err := gmail.NewService(ctx, option.WithTokenSource(src))
	if err != nil {
		return nil, err
	}
	return svc, nil
}

func (a api) connectionData(ctx context.Context) (*vars.Vars, error) {
	cid, err := sdkmodule.FunctionConnectionIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if !cid.IsValid() {
		cid = a.cid // Fallback during authentication flows.
	}

	vs, err := a.vars.Get(ctx, sdktypes.NewVarScopeID(cid))
	if err != nil {
		return nil, err
	}

	var vars vars.Vars
	vs.Decode(&vars)
	return &vars, nil
}

func (a api) oauthTokenSource(ctx context.Context, data string) (oauth2.TokenSource, error) {
	tok, err := sdkintegrations.DecodeOAuthData(data)
	if err != nil {
		return nil, err
	}

	return oauthConfig().TokenSource(ctx, tok.Token), nil
}

// TODO(ENG-112): Use OAuth().Get() instead of calling this function.
func oauthConfig() *oauth2.Config {
	addr := os.Getenv("WEBHOOK_ADDRESS")
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Endpoint:     google.Endpoint,
		RedirectURL:  fmt.Sprintf("https://%s/oauth/redirect/google", addr),
		// https://developers.google.com/gmail/api/auth/scopes
		Scopes: []string{
			// Non-sensitive.
			googleoauth2.OpenIDScope,
			googleoauth2.UserinfoEmailScope,
			googleoauth2.UserinfoProfileScope,
			// Restricted.
			gmail.GmailModifyScope,
			gmail.GmailSettingsBasicScope,
		},
	}
}

func (a api) jwtTokenSource(ctx context.Context, data string) (oauth2.TokenSource, error) {
	scopes := oauthConfig().Scopes

	cfg, err := google.JWTConfigFromJSON([]byte(data), scopes...)
	if err != nil {
		return nil, err
	}

	return cfg.TokenSource(ctx), nil
}
