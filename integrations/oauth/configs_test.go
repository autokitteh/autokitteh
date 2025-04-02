package oauth

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/integrations/auth0"
	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func TestGetConfig(t *testing.T) {
	vs := newFakeVars()
	cid := sdktypes.NewConnectionID()
	vsid := sdktypes.NewVarScopeID(cid)
	require.NoError(t, vs.Set(t.Context(),
		sdktypes.NewVar(auth0.ClientIDVar).SetValue("id").WithScopeID(vsid),
		sdktypes.NewVar(auth0.ClientSecretVar).SetValue("secret").WithScopeID(vsid),
		sdktypes.NewVar(auth0.DomainVar).SetValue("domain").WithScopeID(vsid),

		sdktypes.NewVar(common.PrivateClientIDVar).SetValue("id").WithScopeID(vsid),
		sdktypes.NewVar(common.PrivateClientSecretVar).SetValue("secret").WithScopeID(vsid),
	))

	o := &OAuth{cfg: &Config{}, vars: vs}
	require.NoError(t, o.Start(nil))

	wantAuth0 := deepCopy(o.oauthConfigs["auth0"])
	wantAuth0.Config.ClientID = "id"
	wantAuth0.Config.ClientSecret = "secret"
	wantAuth0.Config.Endpoint.AuthURL = "https://domain/oauth/authorize"
	wantAuth0.Config.Endpoint.DeviceAuthURL = "https://domain/oauth/device/code"
	wantAuth0.Config.Endpoint.TokenURL = "https://domain/oauth/token"
	wantAuth0.Opts["audience"] = "https://domain/api/v2/"

	wantPrivateSlack := deepCopy(o.oauthConfigs["slack"])
	wantPrivateSlack.Config.ClientID = "id"
	wantPrivateSlack.Config.ClientSecret = "secret"

	tests := []struct {
		name        string
		integration string
		cid         sdktypes.ConnectionID
		authType    string
		want        oauthConfig
		wantErr     bool
	}{
		{
			name:        "default_auth0",
			integration: "auth0",
			authType:    integrations.OAuthDefault,
			cid:         cid,
			want:        wantAuth0,
		},
		{
			name:        "default_slack",
			integration: "slack",
			authType:    integrations.OAuthDefault,
			cid:         sdktypes.InvalidConnectionID,
			want:        o.oauthConfigs["slack"],
		},
		{
			name:        "private_auth0",
			integration: "auth0",
			authType:    integrations.OAuthPrivate,
			cid:         cid,
			want:        wantAuth0,
		},
		{
			name:        "private_slack",
			integration: "slack",
			authType:    integrations.OAuthPrivate,
			cid:         cid,
			want:        wantPrivateSlack,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			v := sdktypes.NewVar(common.AuthTypeVar).SetValue(tt.authType)
			require.NoError(t, vs.Set(ctx, v.WithScopeID(vsid)))

			gotConfig, gotOpts, err := o.GetConfig(ctx, tt.integration, tt.cid)
			if (err != nil) != tt.wantErr {
				t.Errorf("OAuth.OAuthConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotConfig, tt.want.Config) {
				t.Errorf("OAuth.OAuthConfig() = %v, want %v", gotConfig, tt.want.Config)
			}
			if !reflect.DeepEqual(gotOpts, tt.want.Opts) {
				t.Errorf("OAuth.OAuthConfig() = %v, want %v", gotOpts, tt.want.Opts)
			}
		})
	}
}

func TestDefaultGHESGitHubURLs(t *testing.T) {
	t.Setenv("GITHUB_APP_NAME", "app-name")
	t.Setenv("GITHUB_CLIENT_ID", "id")
	t.Setenv("GITHUB_CLIENT_SECRET", "secret")
	t.Setenv("GITHUB_ENTERPRISE_URL", "https://github.test.com")

	v := newFakeVars()
	ctx := t.Context()
	cid := sdktypes.NewConnectionID()
	vsid := sdktypes.NewVarScopeID(cid)
	require.NoError(t, v.Set(ctx,
		sdktypes.NewVar(common.AuthTypeVar).SetValue(integrations.OAuthDefault).WithScopeID(vsid),
	))

	o := &OAuth{cfg: &Config{Address: "example.com"}, vars: v}
	require.NoError(t, o.Start(nil))

	cfg, _, err := o.GetConfig(ctx, "github", cid)
	require.NoError(t, err)

	assert.Equal(t, "id", cfg.ClientID)
	assert.Equal(t, "secret", cfg.ClientSecret)
	assert.Equal(t, "https://example.com/oauth/redirect/github", cfg.RedirectURL)

	assert.Equal(t, "https://github.test.com/github-apps/app-name/installations/new", cfg.Endpoint.AuthURL)
	assert.Equal(t, "https://github.test.com/login/device/code", cfg.Endpoint.DeviceAuthURL)
	assert.Equal(t, "https://github.test.com/login/oauth/access_token", cfg.Endpoint.TokenURL)
}

func TestDeepCopy(t *testing.T) {
	o := &OAuth{cfg: &Config{}}
	require.NoError(t, o.Start(nil))

	orig := o.oauthConfigs["slack"]

	copy := deepCopy(orig)
	copy.Config.ClientID = "new_client_id"
	copy.Config.ClientSecret = "new_client_secret"
	copy.Config.Endpoint.AuthURL = "new_auth_url"
	copy.Config.Endpoint.DeviceAuthURL = "new_device_auth_url"
	copy.Config.Endpoint.TokenURL = "new_token_url"
	copy.Config.Endpoint.AuthStyle = oauth2.AuthStyleAutoDetect
	copy.Config.RedirectURL = "new_redirect_url"
	copy.Config.Scopes[0] = "new_scope"
	copy.Opts["new_key"] = "new_value"

	assert.NotEqual(t, orig.Config.ClientID, copy.Config.ClientID)
	assert.NotEqual(t, orig.Config.ClientSecret, copy.Config.ClientSecret)
	assert.NotEqual(t, orig.Config.Endpoint.AuthURL, copy.Config.Endpoint.AuthURL)
	assert.NotEqual(t, orig.Config.Endpoint.DeviceAuthURL, copy.Config.Endpoint.DeviceAuthURL)
	assert.NotEqual(t, orig.Config.Endpoint.TokenURL, copy.Config.Endpoint.TokenURL)
	assert.Equal(t, orig.Config.Endpoint.AuthStyle, oauth2.AuthStyleInHeader)
	assert.NotEqual(t, orig.Config.RedirectURL, copy.Config.RedirectURL)
	assert.NotEqual(t, orig.Config.Scopes, copy.Config.Scopes)
	assert.NotContains(t, orig.Opts, "new_key")
}

func TestPrivatize(t *testing.T) {
	tests := []struct {
		name       string
		vars       sdktypes.Vars
		wantID     string
		wantSecret string
	}{
		{
			name: "default_slack",
			vars: sdktypes.NewVars(
				sdktypes.NewVar(common.AuthTypeVar).SetValue(integrations.OAuthDefault),
			),
			wantID:     "",
			wantSecret: "",
		},
		{
			name: "private_slack",
			vars: sdktypes.NewVars(
				sdktypes.NewVar(common.AuthTypeVar).SetValue(integrations.OAuthPrivate),
				sdktypes.NewVar(common.PrivateClientIDVar).SetValue("id"),
				sdktypes.NewVar(common.PrivateClientSecretVar).SetValue("secret"),
			),
			wantID:     "id",
			wantSecret: "secret",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := newFakeVars()
			ctx := t.Context()
			cid := sdktypes.NewConnectionID()
			vsid := sdktypes.NewVarScopeID(cid)
			vs := tt.vars.WithScopeID(vsid)
			require.NoError(t, v.Set(ctx, vs...))

			o := &OAuth{cfg: &Config{Address: "example.com"}, vars: v}
			require.NoError(t, o.Start(nil))

			cfg, _, err := o.GetConfig(ctx, "slack", cid)
			require.NoError(t, err)

			assert.Equal(t, tt.wantID, cfg.ClientID)
			assert.Equal(t, tt.wantSecret, cfg.ClientSecret)
			assert.Equal(t, "https://example.com/oauth/redirect/slack", cfg.RedirectURL)
		})
	}
}

func TestSetupAuth0(t *testing.T) {
	v := newFakeVars()
	ctx := t.Context()
	cid := sdktypes.NewConnectionID()
	vsid := sdktypes.NewVarScopeID(cid)
	require.NoError(t, v.Set(ctx,
		sdktypes.NewVar(auth0.ClientIDVar).SetValue("id").WithScopeID(vsid),
		sdktypes.NewVar(auth0.ClientSecretVar).SetValue("secret").WithScopeID(vsid),
		sdktypes.NewVar(auth0.DomainVar).SetValue("domain").WithScopeID(vsid),
	))

	o := &OAuth{cfg: &Config{Address: "example.com"}, vars: v}
	require.NoError(t, o.Start(nil))

	cfg, opts, err := o.GetConfig(ctx, "auth0", cid)
	require.NoError(t, err)

	assert.Equal(t, "id", cfg.ClientID)
	assert.Equal(t, "secret", cfg.ClientSecret)
	assert.Equal(t, "https://example.com/oauth/redirect/auth0", cfg.RedirectURL)

	assert.Contains(t, cfg.Endpoint.AuthURL, "https://domain/")
	assert.Contains(t, cfg.Endpoint.DeviceAuthURL, "https://domain/")
	assert.Contains(t, cfg.Endpoint.TokenURL, "https://domain/")
	assert.Contains(t, opts, "audience")
	assert.Contains(t, opts["audience"], "https://domain/")
}

func TestPrivatizeGitHub(t *testing.T) {
	v := newFakeVars()
	ctx := t.Context()
	cid := sdktypes.NewConnectionID()
	vsid := sdktypes.NewVarScopeID(cid)
	require.NoError(t, v.Set(ctx,
		sdktypes.NewVar(common.AuthTypeVar).SetValue(integrations.OAuthPrivate).WithScopeID(vsid),
		sdktypes.NewVar(common.PrivateClientIDVar).SetValue("id").WithScopeID(vsid),
		sdktypes.NewVar(common.PrivateClientSecretVar).SetValue("secret").WithScopeID(vsid),
		sdktypes.NewVar(sdktypes.NewSymbol("app_name")).SetValue("app-name").WithScopeID(vsid),
	))

	o := &OAuth{cfg: &Config{Address: "example.com"}, vars: v}
	require.NoError(t, o.Start(nil))

	cfg, _, err := o.GetConfig(ctx, "github", cid)
	require.NoError(t, err)

	assert.Equal(t, "id", cfg.ClientID)
	assert.Equal(t, "secret", cfg.ClientSecret)
	assert.Equal(t, "https://example.com/oauth/redirect/github", cfg.RedirectURL)

	assert.Equal(t, "https://github.com/apps/app-name/installations/new", cfg.Endpoint.AuthURL)
	assert.Equal(t, "https://github.com/login/device/code", cfg.Endpoint.DeviceAuthURL)
	assert.Equal(t, "https://github.com/login/oauth/access_token", cfg.Endpoint.TokenURL)
}
