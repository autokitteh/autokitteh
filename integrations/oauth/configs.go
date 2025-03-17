package oauth

import (
	"context"
	"fmt"
	"maps"
	"net/url"
	"os"
	"slices"

	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/integrations/auth0"
	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type oauthConfig struct {
	Config *oauth2.Config
	Opts   map[string]string
}

// initConfigs initializes the AutoKitteh server's default OAuth 2.0 configurations
// for all AutoKitteh integrations. This map must not be modified during runtime.
func (o *OAuth) initConfigs() {
	o.oauthConfigs = map[string]oauthConfig{
		// https://auth0.com/docs/api/authentication
		// https://auth0.com/docs/get-started/authentication-and-authorization-flow/authorization-code-flow
		"auth0": {
			Config: &oauth2.Config{
				// Special case: the client ID, secret, and URLs always
				// depend on a connection variable (Auth0 domain); they
				// are not global even in the default OAuth mode.
				Scopes: []string{
					// https://auth0.com/docs/get-started/apis/scopes/openid-connect-scopes
					"openid",
					"profile",
					"email",
					// https://auth0.com/docs/get-started/authentication-and-authorization-flow/authorization-code-flow/add-login-auth-code-flow#post-to-token-url-example
					"offline_access", // For refresh tokens

					"read:users", // For reading user data
					"read:users_app_metadata",
				},
			},
			Opts: map[string]string{
				// https://auth0.com/docs/get-started/applications/application-grant-types
				"grant_type": "client_credentials",
			},
		},

		"confluence": {},

		"discord": {},

		"github": {},

		"gmail": {},

		"google": {},

		"googlecalendar": {},

		"googlechat": {},

		"googledrive": {},

		"googleforms": {},

		"googlesheets": {},

		"height": {},

		"hubspot": {},

		"jira": {},

		"linear": {},

		"microsoft": {},

		"microsoft_teams": {},

		// https://help.salesforce.com/s/articleView?id=xcloud.remoteaccess_oauth_web_server_flow.htm
		"salesforce": {
			Config: &oauth2.Config{
				ClientID:     os.Getenv("SALESFORCE_CLIENT_ID"),
				ClientSecret: os.Getenv("SALESFORCE_CLIENT_SECRET"),
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://login.salesforce.com/services/oauth2/authorize",
					TokenURL: "https://login.salesforce.com/services/oauth2/token",
				},
				// https://help.salesforce.com/s/articleView?id=xcloud.remoteaccess_oauth_tokens_scopes.htm
				Scopes: []string{"api", "refresh_token"},
			},
		},

		// https://api.slack.com/authentication/oauth-v2
		"slack": {
			Config: &oauth2.Config{
				ClientID:     os.Getenv("SLACK_CLIENT_ID"),
				ClientSecret: os.Getenv("SLACK_CLIENT_SECRET"),
				Endpoint: oauth2.Endpoint{
					AuthURL: "https://slack.com/oauth/v2/authorize",
					// https://api.slack.com/methods/oauth.v2.access
					TokenURL: "https://slack.com/api/oauth.v2.access",
					// https://api.slack.com/authentication/oauth-v2#using
					AuthStyle: oauth2.AuthStyleInHeader,
				},
				// https://docs.autokitteh.com/integrations/slack/default_oauth
				// https://api.slack.com/apps/A05F30M6W3H/oauth
				Scopes: []string{
					"app_mentions:read",
					"bookmarks:read",
					"bookmarks:write",
					"channels:history",
					"channels:manage",
					"channels:read",
					"chat:write",
					"chat:write.customize",
					"chat:write.public",
					"commands",
					"dnd:read",
					"groups:history",
					"groups:read",
					"groups:write",
					"im:history",
					"im:read",
					"im:write",
					"mpim:history",
					"mpim:read",
					"mpim:write",
					"reactions:read",
					"reactions:write",
					"users.profile:read",
					"users:read",
					"users:read.email",
				},
			},
		},

		// https://developers.zoom.us/docs/integrations/oauth/
		"zoom": {
			Config: &oauth2.Config{
				ClientID:     os.Getenv("ZOOM_CLIENT_ID"),
				ClientSecret: os.Getenv("ZOOM_CLIENT_SECRET"),
				Endpoint: oauth2.Endpoint{
					AuthURL:       "https://zoom.us/oauth/authorize",
					TokenURL:      "https://zoom.us/oauth/token",
					DeviceAuthURL: "https://zoom.us/oauth/devicecode",
				},
			},
		},
	}

	o.initRedirectURLs()

	// Optional integration-specific customizations.
	o.initAtlassianURLs()
	o.initGitHubURLs()
}

func (o *OAuth) initRedirectURLs() {
	for k, v := range o.oauthConfigs {
		if v.Config == nil {
			continue
		}

		// The redirect URL should be based on the AutoKitteh
		// server's public address and the integration's name.
		u, err := url.JoinPath(o.BaseURL, "/oauth/redirect", k)
		if err != nil {
			o.logger.Error("failed to set OAuth redirect URL",
				zap.String("integration", k), zap.Error(err),
			)
			continue
		}

		// Set the field in the referenced config,
		// don't replace the referenced config.
		v.Config.RedirectURL = u
	}
}

func (o *OAuth) initAtlassianURLs() {
	// TODO: Implement this function.
}

func (o *OAuth) initGitHubURLs() {
	// TODO: Implement this function.
}

// SimpleConfig is a temporary function while we transition from the
// configs in "/internal/backend/oauth/oauth.go" to the new ones here.
// It will be removed once all config data is moved here and tested.
func (o *OAuth) SimpleConfig(integration string) *oauth2.Config {
	c, ok := o.oauthConfigs[integration]
	if !ok {
		o.logger.Error("requested OAuth config not found", zap.String("integration", integration))
		return nil
	}
	if c.Config == nil {
		o.logger.Error("requested OAuth config not set", zap.String("integration", integration))
		return nil
	}
	return c.Config
}

// OAuthConfig returns the OAuth 2.0 configuration for the given integration.
// The connection ID may be nil, but if it's not, this function tries to use private
// OAuth app settings from the connection's variables instead of the integration's defaults.
func (o *OAuth) OAuthConfig(ctx context.Context, integration string, cid sdktypes.ConnectionID) (oauthConfig, error) {
	cfg, ok := o.oauthConfigs[integration]
	if !ok {
		return oauthConfig{}, fmt.Errorf("%w: OAuth ID: %s", sdkerrors.ErrNotFound, integration)
	}

	// No connection ID - return the default configuration.
	if !cid.IsValid() {
		return cfg, nil
	}

	// Failed to read connection variables - return the default configuration.
	vs, err := o.vars.Get(ctx, sdktypes.NewVarScopeID(cid))
	if err != nil {
		o.logger.Error("failed to read connection variables",
			zap.String("integration", integration),
			zap.String("connection_id", cid.String()),
			zap.Error(err),
		)
		return cfg, nil
	}

	// The connection doesn't use private OAuth - return the default configuration.
	// Special exception: Auth0 requires manipulation even in the default OAuth mode.
	if integration != "auth0" && common.ReadAuthType(vs) != integrations.OAuthPrivate {
		return cfg, nil
	}

	// From this point on, we manipulate the config based on the connection's
	// variables, so ensure we no longer reference the server's default.
	cfg = deepCopy(cfg)

	switch integration {
	case "auth0":
		fixAuth0(vs, &cfg)
	case "github":
		privatizeGitHub(vs, &cfg)
	case "microsoft", "microsoft_teams":
		privatize(vs, &cfg)
		privatizeMicrosoft(vs, &cfg)
	default:
		privatize(vs, &cfg)
	}

	return cfg, nil
}

// deepCopy ensures that we don't modify the server's default OAuth 2.0 configurations.
func deepCopy(c oauthConfig) oauthConfig {
	optsCopy := make(map[string]string, len(c.Opts))
	maps.Copy(optsCopy, c.Opts) // maps.Clone won't work because it's shallow.

	return oauthConfig{
		Config: &oauth2.Config{
			ClientID:     c.Config.ClientID,
			ClientSecret: c.Config.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:       c.Config.Endpoint.AuthURL,
				DeviceAuthURL: c.Config.Endpoint.DeviceAuthURL,
				TokenURL:      c.Config.Endpoint.TokenURL,
				AuthStyle:     c.Config.Endpoint.AuthStyle,
			},
			RedirectURL: c.Config.RedirectURL,
			Scopes:      slices.Clone(c.Config.Scopes), // Shallow, but sufficient.
		},
		Opts: optsCopy,
	}
}

// fixAuth0 needs to run even when the connection is using the AutoKitteh server's
// default OAuth app, not just a private one like the [privatize] function.
func fixAuth0(vs sdktypes.Vars, c *oauthConfig) {
	c.Config.ClientID = vs.GetValue(auth0.ClientIDVar)
	c.Config.ClientSecret = vs.GetValue(auth0.ClientSecretVar)

	domain := vs.GetValue(auth0.DomainVar)
	c.Config.Endpoint.AuthURL = fmt.Sprintf("https://%s/oauth/authorize", domain)
	c.Config.Endpoint.DeviceAuthURL = fmt.Sprintf("https://%s/oauth/device/code", domain)
	c.Config.Endpoint.TokenURL = fmt.Sprintf("https://%s/oauth/token", domain)

	c.Opts["audience"] = fmt.Sprintf("https://%s/api/v2/", domain)
}

func privatize(vs sdktypes.Vars, c *oauthConfig) {
	c.Config.ClientID = vs.GetValue(common.PrivateClientIDVar)
	c.Config.ClientSecret = vs.GetValue(common.PrivateClientSecretVar)
}

func privatizeGitHub(vs sdktypes.Vars, c *oauthConfig) {
	// TODO: Implement this function.
}

func privatizeMicrosoft(vs sdktypes.Vars, c *oauthConfig) {
	// TODO(INT-202): Update MS endpoints ("common" --> tenant ID).
}
