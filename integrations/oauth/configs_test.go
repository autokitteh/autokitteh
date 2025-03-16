package oauth

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func TestOAuthConfig(t *testing.T) {
	o := &OAuth{cfg: &Config{}, BaseURL: "https://example.com"}
	require.NoError(t, o.Start(nil))

	tests := []struct {
		name        string
		integration string
		cid         sdktypes.ConnectionID
		want        oauthConfig
		wantErr     bool
	}{
		{
			name:        "default_auth0",
			integration: "auth0",
			cid:         sdktypes.InvalidConnectionID,
			want:        oauthConfig{},
		},
		{
			name:        "default_slack",
			integration: "slack",
			cid:         sdktypes.InvalidConnectionID,
			want:        o.oauthConfigs["slack"],
		},
		// TODO: "private_auth0"
		// TODO: "private_slack"
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			got, err := o.OAuthConfig(ctx, tt.integration, tt.cid)
			if (err != nil) != tt.wantErr {
				t.Errorf("OAuth.OAuthConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OAuth.OAuthConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeepCopy(t *testing.T) {
	o := &OAuth{cfg: &Config{}, BaseURL: "https://example.com"}
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
	copy.Config.Scopes = []string{"new_scope"}
	copy.Opts["new_key"] = "new_value"

	assert.NotEqual(t, copy.Config.ClientID, orig.Config.ClientID)
	assert.NotEqual(t, copy.Config.ClientSecret, orig.Config.ClientSecret)
	assert.NotEqual(t, copy.Config.Endpoint.AuthURL, orig.Config.Endpoint.AuthURL)
	assert.NotEqual(t, copy.Config.Endpoint.DeviceAuthURL, orig.Config.Endpoint.DeviceAuthURL)
	assert.NotEqual(t, copy.Config.Endpoint.TokenURL, orig.Config.Endpoint.TokenURL)
	assert.Equal(t, oauth2.AuthStyleInHeader, orig.Config.Endpoint.AuthStyle)
	assert.NotEqual(t, copy.Config.RedirectURL, orig.Config.RedirectURL)
	assert.NotEqual(t, copy.Config.Scopes, orig.Config.Scopes)
	assert.NotContains(t, orig.Opts, "new_key")
}
