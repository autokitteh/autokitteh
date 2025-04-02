package oauth

import (
	"context"
	"fmt"
	"maps"
	"net/url"
	"os"
	"slices"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/microsoft"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/chat/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/forms/v1"
	"google.golang.org/api/gmail/v1"
	googleoauth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/sheets/v4"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/integrations/auth0"
	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// internalFlags is a central place to define flags for special integration-specific
// handling in this package. These flags are set in [initConfigs] and
// checked with [flags], but never exposed outside this package.
type internalFlags struct {
	customizeDefaultEndpoint bool // Auth0 (used in this file)
	expiryMissingInToken     bool // Salesforce (used in token.go)
	useJWTsNotOAuth          bool // GitHub (used in webhooks.go)
}

type oauthConfig struct {
	Config *oauth2.Config
	Opts   map[string]string
	flags  internalFlags
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
			flags: internalFlags{
				customizeDefaultEndpoint: true,
			},
		},

		// https://developer.atlassian.com/cloud/confluence/oauth-2-3lo-apps/
		"confluence": {
			Config: &oauth2.Config{
				ClientID:     os.Getenv("CONFLUENCE_CLIENT_ID"),
				ClientSecret: os.Getenv("CONFLUENCE_CLIENT_SECRET"),
				// Special case: the addresses in the endpoint URLs may be
				// customized even in the default OAuth mode: see [initAtlassianURLs].
				Scopes: []string{
					"write:confluence-content",
					"read:confluence-space.summary", // Needed?
					"write:confluence-space",
					"write:confluence-file",
					"read:confluence-props", // Needed?
					"write:confluence-props",
					"manage:confluence-configuration",    // Needed?
					"read:confluence-content.all",        // Needed?
					"search:confluence",                  // Needed?
					"read:confluence-content.permission", // Needed?
					"read:confluence-user",
					"read:confluence-groups", // Needed?
					"write:confluence-groups",
					"readonly:content.attachment:confluence", // Needed?
					// User identity API.
					"read:account",
					// https://developer.atlassian.com/cloud/jira/platform/oauth-2-3lo-apps/#use-a-refresh-token-to-get-another-access-token-and-refresh-token-pair
					"offline_access",
				},
			},
		},

		// https://docs.github.com/en/apps/using-github-apps/about-using-github-apps
		"github": {
			Config: &oauth2.Config{
				ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
				ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
				// Special case: the addresses in the endpoint URLs may be
				// customized even in the default OAuth mode: see [initGitHubURLs].
			},
			flags: internalFlags{
				useJWTsNotOAuth: true,
			},
		},

		// https://developers.google.com/gmail/api/auth/scopes
		"gmail": {
			Config: googleConfig([]string{
				// Non-sensitive.
				googleoauth2.OpenIDScope,
				googleoauth2.UserinfoEmailScope,
				googleoauth2.UserinfoProfileScope,
				// Restricted.
				gmail.GmailModifyScope,
				gmail.GmailSettingsBasicScope,
			}),
			Opts: offlineOpts(),
		},

		"google": {
			Config: googleConfig([]string{
				// Non-sensitive.
				googleoauth2.OpenIDScope,
				googleoauth2.UserinfoEmailScope,
				googleoauth2.UserinfoProfileScope,
				drive.DriveFileScope, // See ENG-1701
				// Sensitive.
				calendar.CalendarScope,
				calendar.CalendarEventsScope,
				chat.ChatMembershipsScope,
				chat.ChatMessagesScope,
				chat.ChatSpacesScope,
				forms.FormsBodyScope,
				forms.FormsResponsesReadonlyScope,
				sheets.SpreadsheetsScope,
				// Restricted.
				// drive.DriveScope, // See ENG-1701
				gmail.GmailModifyScope,
				gmail.GmailSettingsBasicScope,
			}),
			Opts: offlineOpts(),
		},

		// https://developers.google.com/calendar/api/auth
		"googlecalendar": {
			Config: googleConfig([]string{
				// Non-sensitive.
				googleoauth2.OpenIDScope,
				googleoauth2.UserinfoEmailScope,
				googleoauth2.UserinfoProfileScope,
				// Sensitive.
				calendar.CalendarScope,
				calendar.CalendarEventsScope,
			}),
			Opts: offlineOpts(),
		},

		// https://developers.google.com/chat/api/guides/auth#chat-api-scopes
		"googlechat": {
			Config: googleConfig([]string{
				// Non-sensitive.
				googleoauth2.OpenIDScope,
				googleoauth2.UserinfoEmailScope,
				googleoauth2.UserinfoProfileScope,
				// Sensitive.
				chat.ChatMembershipsScope,
				chat.ChatMessagesScope,
				chat.ChatSpacesScope,
			}),
			Opts: offlineOpts(),
		},

		// https://developers.google.com/drive/api/guides/api-specific-auth
		"googledrive": {
			Config: googleConfig([]string{
				// Non-sensitive.
				googleoauth2.OpenIDScope,
				googleoauth2.UserinfoEmailScope,
				googleoauth2.UserinfoProfileScope,
				drive.DriveFileScope, // See ENG-1701
				// Restricted.
				// drive.DriveScope, // See ENG-1701
			}),
			Opts: offlineOpts(),
		},

		// https://developers.google.com/identity/protocols/oauth2/scopes#script
		"googleforms": {
			Config: googleConfig([]string{
				// Non-sensitive.
				googleoauth2.OpenIDScope,
				googleoauth2.UserinfoEmailScope,
				googleoauth2.UserinfoProfileScope,
				// Sensitive.
				forms.FormsBodyScope,
				forms.FormsResponsesReadonlyScope,
			}),
			Opts: offlineOpts(),
		},

		// https://developers.google.com/sheets/api/scopes
		"googlesheets": {
			Config: googleConfig([]string{
				// Non-sensitive.
				googleoauth2.OpenIDScope,
				googleoauth2.UserinfoEmailScope,
				googleoauth2.UserinfoProfileScope,
				// Sensitive.
				sheets.SpreadsheetsScope,
			}),
			Opts: offlineOpts(),
		},

		// https://height.notion.site/OAuth-Apps-on-Height-a8ebeab3f3f047e3857bd8ce60c2f640
		"height": {
			Config: &oauth2.Config{
				ClientID:     os.Getenv("HEIGHT_CLIENT_ID"),
				ClientSecret: os.Getenv("HEIGHT_CLIENT_SECRET"),
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://height.app/oauth/authorization",
					TokenURL: "https://api.height.app/oauth/tokens",
				},
				Scopes: []string{"api"},
			},
			Opts: map[string]string{
				// This is a workaround for Height's non-standard OAuth 2.0 flow
				// which expects the scopes string in the exchange request as well.
				"scope": "api",
			},
		},

		// https://developers.hubspot.com/beta-docs/guides/apps/authentication/working-with-oauth
		"hubspot": {
			Config: &oauth2.Config{
				ClientID:     os.Getenv("HUBSPOT_CLIENT_ID"),
				ClientSecret: os.Getenv("HUBSPOT_CLIENT_SECRET"),
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://app.hubspot.com/oauth/authorize",
					TokenURL: "https://api.hubapi.com/oauth/v1/token",
				},
				Scopes: []string{
					"crm.objects.companies.read",
					"crm.objects.companies.write",
					"crm.objects.contacts.read",
					"crm.objects.contacts.write",
					"crm.objects.deals.read",
					"crm.objects.deals.write",
					"crm.objects.owners.read",
				},
			},
		},

		// https://developer.atlassian.com/cloud/jira/platform/oauth-2-3lo-apps/
		"jira": {
			Config: &oauth2.Config{
				ClientID:     os.Getenv("JIRA_CLIENT_ID"),
				ClientSecret: os.Getenv("JIRA_CLIENT_SECRET"),
				// Special case: the addresses in the endpoint URLs may be
				// customized even in the default OAuth mode: see [initAtlassianURLs].
				Scopes: []string{
					"read:jira-work", // Needed?
					"read:jira-user",
					"write:jira-work",
					"manage:jira-webhook",
					// User identity API.
					"read:account",
					// https://developer.atlassian.com/cloud/jira/platform/oauth-2-3lo-apps/#use-a-refresh-token-to-get-another-access-token-and-refresh-token-pair
					"offline_access",
				},
			},
		},

		// https://developers.linear.app/docs/oauth/authentication
		"linear": {
			Config: &oauth2.Config{
				ClientID:     os.Getenv("LINEAR_CLIENT_ID"),
				ClientSecret: os.Getenv("LINEAR_CLIENT_SECRET"),
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://linear.app/oauth/authorize",
					TokenURL: "https://api.linear.app/oauth/token",
				},
				Scopes: []string{"read", "write"},
			},
		},

		// https://learn.microsoft.com/en-us/entra/identity-platform/v2-app-types
		"microsoft": {
			Config: microsoftConfig([]string{
				// Teams.
				"Channel.Create",
				"Channel.Delete.All",
				// "ChannelMember.ReadWrite.All", // Application-only.
				"ChannelMessage.Read.All",
				"ChannelMessage.ReadWrite", // Delegated-only.
				"ChannelMessage.Send",      // Delegated-only.
				"ChannelSettings.ReadWrite.All",
				"Chat.Create",
				"Chat.ManageDeletion.All",
				"Chat.ReadWrite.All",
				"ChatMember.ReadWrite", // Delegated-only.
				// "ChatMember.ReadWrite.All", // App-only, alternative: WhereInstalled.
				// "Group.ReadWrite.All", // Special: allow apps to update messages.
				// "Organization.Read.All", // Optional, for daemon apps to get org info.
				"Team.ReadBasic.All",
				"TeamMember.ReadWrite.All",
				// TODO: TeamsAppInstallation.ReadForXXX? TeamsTab?
				// "Teamwork.Migrate.All", // Application-only.
			}),
			Opts: offlineOpts(),
		},

		// https://learn.microsoft.com/en-us/entra/identity-platform/v2-app-types
		"microsoft_teams": {
			Config: microsoftConfig([]string{
				"Channel.Create",
				"Channel.Delete.All",
				// "ChannelMember.ReadWrite.All", // Application-only.
				"ChannelMessage.Read.All",
				"ChannelMessage.ReadWrite", // Delegated-only.
				"ChannelMessage.Send",      // Delegated-only.
				"ChannelSettings.ReadWrite.All",
				"Chat.Create",
				"Chat.ManageDeletion.All",
				"Chat.ReadWrite.All",
				"ChatMember.ReadWrite", // Delegated-only.
				// "ChatMember.ReadWrite.All", // App-only, alternative: WhereInstalled.
				// "Group.ReadWrite.All", // Special: allow apps to update messages.
				// "Organization.Read.All", // Optional, for daemon apps to get org info.
				"Team.ReadBasic.All",
				"TeamMember.ReadWrite.All",
				// TODO: TeamsAppInstallation.ReadForXXX? TeamsTab?
				// "Teamwork.Migrate.All", // Application-only.
			}),
			Opts: offlineOpts(),
		},

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
			flags: internalFlags{
				expiryMissingInToken: true,
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

// googleConfig sets some common hard-coded OAuth 2.0 configuration
// details in all Google integrations, instead of copy-pasting them.
func googleConfig(scopes []string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Endpoint:     google.Endpoint,
		Scopes:       scopes,
	}
}

// microsoftConfig sets some common hard-coded OAuth 2.0 configuration
// details in all Microsoft integrations, instead of copy-pasting them.
// See: https://learn.microsoft.com/en-us/graph/permissions-overview,
// https://learn.microsoft.com/en-us/graph/permissions-reference
func microsoftConfig(scopes []string) *oauth2.Config {
	scopes = append(scopes,
		// Non-sensitive delegated-only permissions:
		// https://learn.microsoft.com/en-us/entra/identity-platform/scopes-oidc#openid-connect-scopes
		"email", "offline_access", "openid", "profile",
		"User.Read",

		// Admin consent required, but important for many operations.
		"User.ReadBasic.All",
	)
	return &oauth2.Config{
		ClientID:     os.Getenv("MICROSOFT_CLIENT_ID"),
		ClientSecret: os.Getenv("MICROSOFT_CLIENT_SECRET"),
		// TODO(INT-202): Update MS endpoints ("common" --> tenant ID).
		Endpoint: microsoft.AzureADEndpoint("common"),
		Scopes:   scopes,
	}
}

// offlineOpts sets some common hard-coded OAuth 2.0 configuration details
// in all Google and Microsoft integrations, instead of copy-pasting them.
func offlineOpts() map[string]string {
	return map[string]string{
		"access_type": "offline", // oauth2.AccessTypeOffline
	}
}

func (o *OAuth) initRedirectURLs() {
	for k, v := range o.oauthConfigs {
		if v.Config == nil {
			continue
		}

		// Special case: Microsoft integrations all use "microsoft".
		if strings.HasPrefix(k, "microsoft") {
			k = "microsoft"
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

const defaultAtlassianBaseURL = "https://api.atlassian.com"

// initAtlassianURLs initializes the endpoint URLs of Confluence and
// Jira, to support on-prem Atlassian servers in the default OAuth mode
// (see also: https://auth.atlassian.com/.well-known/openid-configuration).
func (o *OAuth) initAtlassianURLs() {
	baseURL := os.Getenv("ATLASSIAN_BASE_URL")

	var err error
	baseURL, err = normalizeAddress(baseURL, defaultAtlassianBaseURL)
	if err != nil {
		o.logger.Error("invalid Atlassian base URL", zap.Error(err))
		baseURL = defaultPublicBackendBaseURL
	}

	baseURL = strings.Replace(baseURL, "api", "auth", 1)

	for _, name := range []string{"confluence", "jira"} {
		c := o.oauthConfigs[name].Config
		c.Endpoint.AuthURL = baseURL + "/authorize"
		c.Endpoint.DeviceAuthURL = baseURL + "/oauth/device/code"
		c.Endpoint.TokenURL = baseURL + "/oauth/token"
	}
}

const (
	defaultGitHubAppName = "unknown-app-name"
	defaultGitHubBaseURL = "https://github.com"
)

// initGitHubURLs initializes GitHub's endpoint URLs, to support GitHub Enterprise
// Server (GHES, i.e. on-prem) in the default OAuth mode. See also [privatizeGitHub].
func (o *OAuth) initGitHubURLs() {
	baseURL := os.Getenv("GITHUB_ENTERPRISE_URL")

	appsDir := "apps" // github.com
	if baseURL != "" {
		appsDir = "github-apps" // GHES
	}

	appName := os.Getenv("GITHUB_APP_NAME")
	if appName == "" {
		appName = defaultGitHubAppName
	}

	var err error
	baseURL, err = normalizeAddress(baseURL, defaultGitHubBaseURL)
	if err != nil {
		o.logger.Error("invalid GitHub base URL", zap.Error(err))
		baseURL = defaultPublicBackendBaseURL
	}

	c := o.oauthConfigs["github"].Config
	// https://docs.github.com/en/apps/using-github-apps/installing-a-github-app-from-a-third-party#installing-a-github-app
	c.Endpoint.AuthURL = fmt.Sprintf("%s/%s/%s/installations/new", baseURL, appsDir, appName)
	// https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/authorizing-oauth-apps#device-flow
	c.Endpoint.DeviceAuthURL = baseURL + "/login/device/code"
	// https://docs.github.com/en/enterprise-server/apps/oauth-apps/building-oauth-apps/authorizing-oauth-apps#2-users-are-redirected-back-to-your-site-by-github
	// https://docs.github.com/en/enterprise-server/apps/sharing-github-apps/making-your-github-app-available-for-github-enterprise-server#the-app-code-must-use-the-correct-urls
	c.Endpoint.TokenURL = baseURL + "/login/oauth/access_token"
	// https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/authorizing-oauth-apps#3-use-the-access-token-to-access-the-api
	c.Endpoint.AuthStyle = oauth2.AuthStyleInHeader
}

func (o *OAuth) flags(integration string) internalFlags {
	return o.oauthConfigs[integration].flags
}

// GetConfig returns the OAuth 2.0 configuration for the given integration.
// The connection ID may be nil, but if it's not, this function tries to use private
// OAuth app settings from the connection's variables instead of the integration's defaults.
func (o *OAuth) GetConfig(ctx context.Context, integration string, cid sdktypes.ConnectionID) (*oauth2.Config, map[string]string, error) {
	cfg, ok := o.oauthConfigs[integration]
	if !ok {
		return nil, nil, fmt.Errorf("%w: OAuth ID: %s", sdkerrors.ErrNotFound, integration)
	}

	// No connection ID - return the default configuration.
	if !cid.IsValid() {
		return cfg.Config, cfg.Opts, nil
	}

	// Failed to read connection variables - return the default configuration.
	vs, err := o.vars.Get(ctx, sdktypes.NewVarScopeID(cid))
	if err != nil {
		o.logger.Error("failed to read connection variables",
			zap.String("integration", integration),
			zap.String("connection_id", cid.String()),
			zap.Error(err),
		)
		return cfg.Config, cfg.Opts, nil
	}

	// The connection doesn't use private OAuth - return the default configuration.
	// Special case: Auth0 requires manipulation even in the default OAuth mode.
	if !o.flags(integration).customizeDefaultEndpoint && common.ReadAuthType(vs) != integrations.OAuthPrivate {
		return cfg.Config, cfg.Opts, nil
	}

	// From this point on, we manipulate the config based on the connection's
	// variables, so ensure we no longer reference the server's default.
	cfg = deepCopy(cfg)

	privatizeClient(vs, &cfg)

	switch integration {
	case "auth0":
		setupAuth0(vs, &cfg)
	case "github":
		privatizeGitHub(vs, &cfg)
	case "microsoft", "microsoft_teams":
		privatizeMicrosoft(vs, &cfg)
	}

	return cfg.Config, cfg.Opts, nil
}

// deepCopy ensures that we don't modify the server's default OAuth 2.0 configurations.
func deepCopy(cfg oauthConfig) oauthConfig {
	optsCopy := make(map[string]string, len(cfg.Opts))
	maps.Copy(optsCopy, cfg.Opts) // maps.Clone won't work because it's shallow.

	return oauthConfig{
		Config: &oauth2.Config{
			ClientID:     cfg.Config.ClientID,
			ClientSecret: cfg.Config.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:       cfg.Config.Endpoint.AuthURL,
				DeviceAuthURL: cfg.Config.Endpoint.DeviceAuthURL,
				TokenURL:      cfg.Config.Endpoint.TokenURL,
				AuthStyle:     cfg.Config.Endpoint.AuthStyle,
			},
			RedirectURL: cfg.Config.RedirectURL,
			Scopes:      slices.Clone(cfg.Config.Scopes), // Shallow, but sufficient.
		},
		Opts: optsCopy,
	}
}

func privatizeClient(vs sdktypes.Vars, cfg *oauthConfig) {
	cfg.Config.ClientID = vs.GetValue(common.PrivateClientIDVar)
	if cfg.Config.ClientID == "" {
		cfg.Config.ClientID = vs.GetValueByString("client_id") // Deprecate this.
	}

	cfg.Config.ClientSecret = vs.GetValue(common.PrivateClientSecretVar)
	if cfg.Config.ClientSecret == "" {
		cfg.Config.ClientSecret = vs.GetValueByString("client_secret") // Deprecate this.
	}
}

// setupAuth0 needs to run even when the connection is using the AutoKitteh server's
// default OAuth app, not just a private one like the [privatize] function.
func setupAuth0(vs sdktypes.Vars, cfg *oauthConfig) {
	domain := vs.GetValue(auth0.DomainVar)
	cfg.Config.Endpoint.AuthURL = fmt.Sprintf("https://%s/oauth/authorize", domain)
	cfg.Config.Endpoint.DeviceAuthURL = fmt.Sprintf("https://%s/oauth/device/code", domain)
	cfg.Config.Endpoint.TokenURL = fmt.Sprintf("https://%s/oauth/token", domain)

	cfg.Opts["audience"] = fmt.Sprintf("https://%s/api/v2/", domain)
}

// TODO(INT-349): Support GHES in private OAuth mode too.
func privatizeGitHub(vs sdktypes.Vars, cfg *oauthConfig) {
	defaultAppName := os.Getenv("GITHUB_APP_NAME")
	if defaultAppName == "" {
		defaultAppName = defaultGitHubAppName
	}
	defaultAppName = fmt.Sprintf("/%s/", defaultAppName)

	authURL := cfg.Config.Endpoint.AuthURL
	// TODO(INT-353): Make GitHub var symbols non-internal, and reuse here.
	privateAppName := fmt.Sprintf("/%s/", vs.GetValueByString("app_name"))
	cfg.Config.Endpoint.AuthURL = strings.Replace(authURL, defaultAppName, privateAppName, 1)
}

func privatizeMicrosoft(vs sdktypes.Vars, c *oauthConfig) {
	// TODO(INT-202): Update MS endpoints ("common" --> tenant ID).
}
