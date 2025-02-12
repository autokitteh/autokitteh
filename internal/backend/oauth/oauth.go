package oauth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
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
	"go.autokitteh.dev/autokitteh/integrations/github"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type oauth struct {
	logger *zap.Logger
	vars   sdkservices.Vars

	// Configs and opts store registration data together.
	// If we replace these in-memory maps with persistent
	// storage, we will merge them into a single table.
	configs map[string]*oauth2.Config
	opts    map[string]map[string]string
}

// intgSetup contains the parameters for configuring an integration's OAuth
type intgSetup struct {
	intg string
	cid  sdktypes.ConnectionID
	cfg  *oauth2.Config
	opts map[string]string
	vars sdktypes.Vars
}

func New(l *zap.Logger, vars sdkservices.Vars) sdkservices.OAuth {
	// TODO(ENG-112): Remove (see Register below).
	redirectURL := fmt.Sprintf("https://%s/oauth/redirect/", os.Getenv("WEBHOOK_ADDRESS"))

	// Determine Atlassian base URL (to support Confluence and Jira on-prem).
	// TODO(ENG-965): From new-connection form instead of env var.
	atlassianBaseURL := os.Getenv("ATLASSIAN_BASE_URL")
	if atlassianBaseURL == "" {
		atlassianBaseURL = "https://api.atlassian.com"
	}
	var err error
	atlassianBaseURL, err = kittehs.NormalizeURL(atlassianBaseURL, true)
	if err != nil {
		l.Fatal("Invalid environment variable value",
			zap.String("name", "ATLASSIAN_BASE_URL"),
			zap.Error(err),
		)
	}
	atlassianBaseURL = strings.Replace(atlassianBaseURL, "api", "auth", 1)
	l.Debug("Atlassian base URL for OAuth", zap.String("url", atlassianBaseURL))

	// Determine GitHub base URL (to support GitHub Enterprise Server, i.e. on-prem).
	// TODO(INT-194): Add support for custom OAuth.
	githubBaseURL := os.Getenv("GITHUB_ENTERPRISE_URL")
	if githubBaseURL == "" {
		githubBaseURL = "https://github.com"
	}
	githubBaseURL, err = kittehs.NormalizeURL(githubBaseURL, true)
	if err != nil {
		l.Fatal("Invalid environment variable value",
			zap.String("name", "GITHUB_ENTERPRISE_URL"),
			zap.Error(err),
		)
	}
	l.Debug("GitHub base URL for OAuth", zap.String("url", githubBaseURL))

	appsDir := "apps"
	if os.Getenv("GITHUB_ENTERPRISE_URL") != "" {
		appsDir = "github-apps"
	}

	return &oauth{
		logger: l,
		vars:   vars,
		// TODO(ENG-112): Construct the following 2 maps with dynamic integration
		// registrations, where each integration registration will call Register
		// below (if it uses OAuth). This hard-coding is EXTREMELY TEMPORARY!
		configs: map[string]*oauth2.Config{
			"auth0": {
				// Auth0 is a special case: environment variables are not supported.
				// All authentication credentials must be stored in `vars`.
				ClientID:     "",
				ClientSecret: "",
				Endpoint: oauth2.Endpoint{
					// Each Auth0 app has a unique domain stored in `vars`.
					// This domain is dynamically replaced during the authentication flow.
					AuthURL:  "https://{{AUTH0_DOMAIN}}/oauth/authorize",
					TokenURL: "https://{{AUTH0_DOMAIN}}/oauth/token",
				},
				RedirectURL: redirectURL + "auth0",
				Scopes: []string{
					"openid",
					"profile",
					"email",
					"read:users", // For reading user data
					"read:users_app_metadata",
					"offline_access", // For refresh tokens
				},
			},
			"confluence": {
				// TODO(ENG-965): From new-connection form instead of env vars.
				ClientID:     os.Getenv("CONFLUENCE_CLIENT_ID"),
				ClientSecret: os.Getenv("CONFLUENCE_CLIENT_SECRET"),
				// https://developer.atlassian.com/cloud/confluence/oauth-2-3lo-apps/
				// https://auth.atlassian.com/.well-known/openid-configuration
				Endpoint: oauth2.Endpoint{
					AuthURL:       atlassianBaseURL + "/authorize",
					TokenURL:      atlassianBaseURL + "/oauth/token",
					DeviceAuthURL: atlassianBaseURL + "/oauth/device/code",
				},
				RedirectURL: redirectURL + "confluence",
				// https://developer.atlassian.com/cloud/confluence/scopes-for-oauth-2-3LO-and-forge-apps/
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

			// Based on:
			// https://github.com/organizations/autokitteh/settings/apps/autokitteh
			"github": {
				ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
				ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
				Endpoint: oauth2.Endpoint{
					// https://docs.github.com/en/apps/using-github-apps/installing-a-github-app-from-a-third-party#installing-a-github-app
					AuthURL: fmt.Sprintf("%s/%s/{{GITHUB_APP_NAME}}/installations/new", githubBaseURL, appsDir),
					// https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/authorizing-oauth-apps#device-flow
					DeviceAuthURL: githubBaseURL + "/login/device/code",
					// https://docs.github.com/en/enterprise-server/apps/oauth-apps/building-oauth-apps/authorizing-oauth-apps#2-users-are-redirected-back-to-your-site-by-github
					// https://docs.github.com/en/enterprise-server/apps/sharing-github-apps/making-your-github-app-available-for-github-enterprise-server#the-app-code-must-use-the-correct-urls
					TokenURL: githubBaseURL + "/login/oauth/access_token",
					// https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/authorizing-oauth-apps#3-use-the-access-token-to-access-the-api
					AuthStyle: oauth2.AuthStyleInHeader,
				},
				RedirectURL: redirectURL + "github", // TODO(ENG-112): Remove (see Register below).
			},

			// Based on:
			// https://console.cloud.google.com/apis/credentials/consent/edit
			"gmail": {
				ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
				ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
				Endpoint:     google.Endpoint,
				RedirectURL:  redirectURL + "google", // TODO(ENG-112): Remove (see Register below).
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
			},

			// Based on:
			// https://console.cloud.google.com/apis/credentials/consent/edit
			"google": {
				ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
				ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
				Endpoint:     google.Endpoint,
				RedirectURL:  redirectURL + "google", // TODO(ENG-112): Remove (see Register below).
				// https://developers.google.com/calendar/api/auth
				// https://developers.google.com/chat/api/guides/auth#chat-api-scopes
				// https://developers.google.com/drive/api/guides/api-specific-auth
				// https://developers.google.com/identity/protocols/oauth2/scopes#script
				// https://developers.google.com/gmail/api/auth/scopes
				// https://developers.google.com/sheets/api/scopes
				Scopes: []string{
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
				},
			},

			// Based on:
			// https://console.cloud.google.com/apis/credentials/consent/edit
			"googlecalendar": {
				ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
				ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
				Endpoint:     google.Endpoint,
				RedirectURL:  redirectURL + "google", // TODO(ENG-112): Remove (see Register below).
				// https://developers.google.com/calendar/api/auth
				Scopes: []string{
					// Non-sensitive.
					googleoauth2.OpenIDScope,
					googleoauth2.UserinfoEmailScope,
					googleoauth2.UserinfoProfileScope,
					// Sensitive.
					calendar.CalendarScope,
					calendar.CalendarEventsScope,
				},
			},

			// Based on:
			// https://console.cloud.google.com/apis/credentials/consent/edit
			"googlechat": {
				ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
				ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
				Endpoint:     google.Endpoint,
				RedirectURL:  redirectURL + "google", // TODO(ENG-112): Remove (see Register below).
				// https://developers.google.com/chat/api/guides/auth#chat-api-scopes
				Scopes: []string{
					// Non-sensitive.
					googleoauth2.OpenIDScope,
					googleoauth2.UserinfoEmailScope,
					googleoauth2.UserinfoProfileScope,
					// Sensitive.
					chat.ChatMembershipsScope,
					chat.ChatMessagesScope,
					chat.ChatSpacesScope,
				},
			},

			// Based on:
			// https://console.cloud.google.com/apis/credentials/consent/edit
			"googledrive": {
				ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
				ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
				Endpoint:     google.Endpoint,
				RedirectURL:  redirectURL + "google", // TODO(ENG-112): Remove (see Register below).
				// https://developers.google.com/drive/api/guides/api-specific-auth
				Scopes: []string{
					// Non-sensitive.
					googleoauth2.OpenIDScope,
					googleoauth2.UserinfoEmailScope,
					googleoauth2.UserinfoProfileScope,
					drive.DriveFileScope, // See ENG-1701
					// Restricted.
					// drive.DriveScope, // See ENG-1701
				},
			},

			// Based on:
			// https://console.cloud.google.com/apis/credentials/consent/edit
			"googleforms": {
				ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
				ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
				Endpoint:     google.Endpoint,
				RedirectURL:  redirectURL + "google", // TODO(ENG-112): Remove (see Register below).
				// https://developers.google.com/identity/protocols/oauth2/scopes#script
				Scopes: []string{
					// Non-sensitive.
					googleoauth2.OpenIDScope,
					googleoauth2.UserinfoEmailScope,
					googleoauth2.UserinfoProfileScope,
					// Sensitive.
					forms.FormsBodyScope,
					forms.FormsResponsesReadonlyScope,
				},
			},

			// Based on:
			// https://console.cloud.google.com/apis/credentials/consent/edit
			"googlesheets": {
				ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
				ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
				Endpoint:     google.Endpoint,
				RedirectURL:  redirectURL + "google", // TODO(ENG-112): Remove (see Register below).
				// https://developers.google.com/sheets/api/scopes
				// TODO: Pass the desired scopes as a runtime argument, to reuse across Google APIs.
				Scopes: []string{
					// Non-sensitive.
					googleoauth2.OpenIDScope,
					googleoauth2.UserinfoEmailScope,
					googleoauth2.UserinfoProfileScope,
					// Sensitive.
					sheets.SpreadsheetsScope,
				},
			},

			// Based on:
			// https://height.notion.site/OAuth-Apps-on-Height-a8ebeab3f3f047e3857bd8ce60c2f640
			"height": {
				ClientID:     os.Getenv("HEIGHT_CLIENT_ID"),
				ClientSecret: os.Getenv("HEIGHT_CLIENT_SECRET"),
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://height.app/oauth/authorization",
					TokenURL: "https://api.height.app/oauth/tokens",
				},
				RedirectURL: redirectURL + "height",
				Scopes: []string{
					"api",
				},
			},

			// Based on:
			// https://developers.hubspot.com/beta-docs/guides/apps/authentication/working-with-oauth
			"hubspot": {
				ClientID:     os.Getenv("HUBSPOT_CLIENT_ID"),
				ClientSecret: os.Getenv("HUBSPOT_CLIENT_SECRET"),
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://app.hubspot.com/oauth/authorize",
					TokenURL: "https://api.hubapi.com/oauth/v1/token",
				},
				RedirectURL: redirectURL + "hubspot",
				Scopes: []string{
					"crm.objects.contacts.read",
					"crm.objects.contacts.write",
					"crm.objects.companies.read",
					"crm.objects.companies.write",
					"crm.objects.deals.read",
					"crm.objects.deals.write",
					"crm.objects.owners.read",
				},
			},

			"jira": {
				// TODO(ENG-965): From new-connection form instead of env vars.
				ClientID:     os.Getenv("JIRA_CLIENT_ID"),
				ClientSecret: os.Getenv("JIRA_CLIENT_SECRET"),
				// https://developer.atlassian.com/cloud/jira/platform/oauth-2-3lo-apps/
				// https://auth.atlassian.com/.well-known/openid-configuration
				Endpoint: oauth2.Endpoint{
					AuthURL:       atlassianBaseURL + "/authorize",
					TokenURL:      atlassianBaseURL + "/oauth/token",
					DeviceAuthURL: atlassianBaseURL + "/oauth/device/code",
				},
				RedirectURL: redirectURL + "jira",
				// https://developer.atlassian.com/cloud/jira/platform/scopes-for-oauth-2-3LO-and-forge-apps/
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

			// Based on: https://developers.linear.app/docs/oauth/authentication
			"linear": {
				ClientID:     os.Getenv("LINEAR_CLIENT_ID"),
				ClientSecret: os.Getenv("LINEAR_CLIENT_SECRET"),
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://linear.app/oauth/authorize",
					TokenURL: "https://api.linear.app/oauth/token",
				},
				RedirectURL: redirectURL + "linear",
				Scopes:      []string{"read", "write"},
			},

			// Based on:
			// https://learn.microsoft.com/en-us/entra/identity-platform/v2-app-types
			"microsoft": {
				ClientID:     os.Getenv("MICROSOFT_CLIENT_ID"),
				ClientSecret: os.Getenv("MICROSOFT_CLIENT_SECRET"),
				Endpoint:     microsoft.AzureADEndpoint("common"),
				RedirectURL:  redirectURL + "microsoft",
				// https://learn.microsoft.com/en-us/graph/permissions-overview
				// https://learn.microsoft.com/en-us/graph/permissions-reference
				// https://learn.microsoft.com/en-us/entra/identity-platform/scopes-oidc#openid-connect-scopes
				Scopes: []string{
					// Non-sensitive delegated-only permissions.
					"email", "offline_access", "openid", "profile",
					"User.Read",

					// Admin consent required, but important for many operations.
					"User.ReadBasic.All",

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
				},
			},

			// Based on:
			// https://learn.microsoft.com/en-us/entra/identity-platform/v2-app-types
			"microsoft_teams": {
				ClientID:     os.Getenv("MICROSOFT_CLIENT_ID"),
				ClientSecret: os.Getenv("MICROSOFT_CLIENT_SECRET"),
				Endpoint:     microsoft.AzureADEndpoint("common"),
				RedirectURL:  redirectURL + "microsoft",
				// https://learn.microsoft.com/en-us/graph/permissions-overview
				// https://learn.microsoft.com/en-us/graph/permissions-reference
				// https://learn.microsoft.com/en-us/entra/identity-platform/scopes-oidc#openid-connect-scopes
				Scopes: []string{
					// Non-sensitive delegated-only permissions.
					"email", "offline_access", "openid", "profile",
					"User.Read",

					// Admin consent required, but important for many operations.
					"User.ReadBasic.All",

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
				},
			},
			"salesforce": {
				// Salesforce is a special case: environment variables are not supported.
				// All authentication credentials must be stored in `vars`.
				ClientID:     "",
				ClientSecret: "",
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://login.salesforce.com/services/oauth2/authorize",
					TokenURL: "https://login.salesforce.com/services/oauth2/token",
				},
				RedirectURL: redirectURL + "salesforce",
				Scopes:      []string{"full"},
			},

			// Based on:
			// https://api.slack.com/apps/A05F30M6W3H
			"slack": {
				ClientID:     os.Getenv("SLACK_CLIENT_ID"),
				ClientSecret: os.Getenv("SLACK_CLIENT_SECRET"),
				Endpoint: oauth2.Endpoint{
					// https://api.slack.com/authentication/oauth-v2
					AuthURL: "https://slack.com/oauth/v2/authorize",
					// https://api.slack.com/methods/oauth.v2.access
					TokenURL: "https://slack.com/api/oauth.v2.access",
					// https://api.slack.com/authentication/oauth-v2#using
					AuthStyle: oauth2.AuthStyleInHeader,
				},
				RedirectURL: redirectURL + "slack", // TODO(ENG-112): Remove (see Register below).
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

			// Based on: https://developers.zoom.us/docs/integrations/oauth/
			"zoom": {
				ClientID:     os.Getenv("ZOOM_CLIENT_ID"),
				ClientSecret: os.Getenv("ZOOM_CLIENT_SECRET"),
				Endpoint: oauth2.Endpoint{
					AuthURL:       "https://zoom.us/oauth/authorize",
					TokenURL:      "https://zoom.us/oauth/token",
					DeviceAuthURL: "https://zoom.us/oauth/devicecode",
				},
				RedirectURL: redirectURL + "zoom",
				// https://developers.zoom.us/docs/integrations/oauth-scopes/
				Scopes: []string{
					"calendar:write",
					"contact:read",
					"meeting:write",
					"meeting_summary:read",
					"recording:read",
					"scheduler:write",
					"user:read",
					"whiteboard:read",
				},
			},
		},

		opts: map[string]map[string]string{
			"auth0": {
				// Using template-style placeholder for AUTH0_DOMAIN
				"audience":   "https://{{AUTH0_DOMAIN}}/api/v2/",
				"grant_type": "client_credentials",
			},
			"gmail": {
				"access_type": "offline", // oauth2.AccessTypeOffline
				"prompt":      "consent", // oauth2.ApprovalForce
			},
			"google": {
				"access_type": "offline", // oauth2.AccessTypeOffline
				"prompt":      "consent", // oauth2.ApprovalForce
			},
			"googlecalendar": {
				"access_type": "offline", // oauth2.AccessTypeOffline
				"prompt":      "consent", // oauth2.ApprovalForce
			},
			"googlechat": {
				"access_type": "offline", // oauth2.AccessTypeOffline
				"prompt":      "consent", // oauth2.ApprovalForce
			},
			"googledrive": {
				"access_type": "offline", // oauth2.AccessTypeOffline
				"prompt":      "consent", // oauth2.ApprovalForce
			},
			"googleforms": {
				"access_type": "offline", // oauth2.AccessTypeOffline
				"prompt":      "consent", // oauth2.ApprovalForce
			},
			"googlesheets": {
				"access_type": "offline", // oauth2.AccessTypeOffline
				"prompt":      "consent", // oauth2.ApprovalForce
			},
			"height": {
				// This is a workaround for Height's non-standard OAuth 2.0 flow
				// which expects the scopes string in the exchange request as well.
				"scope": "api",
			},
			"linear": {
				"prompt": "consent", // oauth2.ApprovalForce
			},
			"microsoft": {
				"access_type": "offline", // oauth2.AccessTypeOffline
				"prompt":      "consent", // oauth2.ApprovalForce
			},
			"microsoft_teams": {
				"access_type": "offline", // oauth2.AccessTypeOffline
				"prompt":      "consent", // oauth2.ApprovalForce
			},
		},
	}
}

func (o *oauth) Register(ctx context.Context, intg string, cfg *oauth2.Config, opts map[string]string) error {
	cfg.RedirectURL = fmt.Sprintf("https://%s/oauth/redirect/%s", os.Getenv("WEBHOOK_ADDRESS"), intg)
	o.configs[intg] = cfg
	o.opts[intg] = opts
	return nil
}

func (o *oauth) Get(ctx context.Context, intg string) (*oauth2.Config, map[string]string, error) {
	cfg, ok := o.configs[intg]
	if !ok {
		return nil, nil, fmt.Errorf("%w: %q not registered as an OAuth ID", sdkerrors.ErrNotFound, intg)
	}
	return cfg, o.opts[intg], nil
}

// applyIntegrationConfig customizes the OAuth2 configuration and options for a specific integration.
// It adjusts the AuthURL, TokenURL, and other parameters based on the integration's setup.
// If custom OAuth is used (instead of the default app), it retrieves and applies the custom configuration.
func (o *oauth) applyIntegrationConfig(ctx context.Context, s *intgSetup) error {
	switch s.intg {
	case "auth0":
		placeholder := "{{AUTH0_DOMAIN}}"
		domain := s.vars.GetValueByString("auth0_domain")

		s.cfg.Endpoint.AuthURL = strings.Replace(s.cfg.Endpoint.AuthURL, placeholder, domain, 1)
		s.cfg.Endpoint.TokenURL = strings.Replace(s.cfg.Endpoint.TokenURL, placeholder, domain, 1)
		s.opts["audience"] = strings.Replace(s.opts["audience"], placeholder, domain, 1)

	case "github":
		placeholder := "{{GITHUB_APP_NAME}}"
		var appName string
		if o.isCustomOAuth(s.vars) {
			// Get app name using appID for custom OAuth
			appID := s.vars.GetValueByString("app_id")
			aid, err := strconv.ParseInt(appID, 10, 64)
			if err != nil {
				return err
			}
			gh, err := github.NewClientFromGitHubAppID(aid, s.vars.GetValueByString("private_key"))
			if err != nil {
				return err
			}
			app, _, err := gh.Apps.Get(ctx, "")
			if err != nil {
				return err
			}
			appName = *app.Slug
		} else {
			// Use default app name from env
			appName = os.Getenv("GITHUB_APP_NAME")
		}
		s.cfg.Endpoint.AuthURL = strings.Replace(s.cfg.Endpoint.AuthURL, placeholder, appName, 1)

	// case "height":

	case "microsoft", "microsoft_teams":
		if o.isCustomOAuth(s.vars) {
			s.cfg.ClientID = s.vars.GetValueByString("private_client_id")
			s.cfg.ClientSecret = s.vars.GetValueByString("private_client_secret")
		}
	}

	return nil
}

func (o *oauth) getConfigWithConnection(ctx context.Context, intg string, cid sdktypes.ConnectionID) (*oauth2.Config, map[string]string, error) {
	if !cid.IsValid() {
		return nil, nil, errors.New("invalid connection ID")
	}

	baseCfg, opts, err := o.Get(ctx, intg)
	if err != nil {
		return nil, nil, err
	}

	cfgCopy := &oauth2.Config{
		ClientID:     baseCfg.ClientID,
		ClientSecret: baseCfg.ClientSecret,
		Endpoint:     baseCfg.Endpoint,
		RedirectURL:  baseCfg.RedirectURL,
		Scopes:       append([]string{}, baseCfg.Scopes...),
	}

	optsCopy := make(map[string]string, len(opts))
	for k, v := range opts {
		optsCopy[k] = v
	}

	vs, err := o.vars.Get(ctx, sdktypes.NewVarScopeID(cid))
	if err != nil {
		return nil, nil, err
	}

	// TODO(INT-202): Update MS endpoints ("common" --> tenant ID).
	if o.isCustomOAuth(vs) {
		cfgCopy.ClientID = vs.GetValueByString("client_id")
		if cfgCopy.ClientID == "" {
			cfgCopy.ClientID = vs.GetValueByString("private_client_id")
		}
		cfgCopy.ClientSecret = vs.GetValueByString("client_secret")
		if cfgCopy.ClientSecret == "" {
			cfgCopy.ClientSecret = vs.GetValueByString("private_client_secret")
		}
	}

	s := &intgSetup{
		intg: intg,
		cid:  cid,
		cfg:  cfgCopy,
		opts: optsCopy,
		vars: vs,
	}

	if err := o.applyIntegrationConfig(ctx, s); err != nil {
		return nil, nil, err
	}

	return cfgCopy, optsCopy, nil
}

func (o *oauth) StartFlow(ctx context.Context, intg string, cid sdktypes.ConnectionID, origin string) (string, error) {
	cfg, opts, err := o.getConfigWithConnection(ctx, intg, cid)
	if err != nil {
		return "", err
	}

	if origin == "" {
		return "", errors.New("missing origin")
	}

	// Identify the relevant connection when we get an OAuth response.
	state := strings.Replace(cid.String(), "con_", "", 1) + "_" + origin

	return cfg.AuthCodeURL(state, authCode(opts)...), nil
}

func (o *oauth) Exchange(ctx context.Context, integration string, cid sdktypes.ConnectionID, code string) (*oauth2.Token, error) {
	// Convert the received temporary authorization code into a refresh token / user access token.
	// ATTENTION: This method may be called by an UNAUTHENTICATED webhook handler, however we CAN
	// use the system user (SetAuthnSystemUser) securely, because only the 3rd-party OAuth
	// provider can provide us a valid "code" parameter related to the connection ID. In other
	// words, if the "code" is fake, or the CID is valid-but-unrelated to the beginning of the
	// OAuth flow, the OAuth provider will reject the exchange, so data leakage is not possible.
	cfg, opts, err := o.getConfigWithConnection(authcontext.SetAuthnSystemUser(ctx), integration, cid)
	if err != nil {
		return nil, fmt.Errorf("bad oauth integration name: %w", err)
	}

	hc := &http.Client{Timeout: exchangeTimeout}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, hc)
	token, err := cfg.Exchange(ctx, code, authCode(opts)...)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	return token, nil
}

func authCode(opts map[string]string) []oauth2.AuthCodeOption {
	var acos []oauth2.AuthCodeOption
	for k, v := range opts {
		acos = append(acos, oauth2.SetAuthURLParam(k, v))
	}
	return acos
}

// Determines if the connection uses custom OAuth based on the presence of a client secret in vars.
func (o *oauth) isCustomOAuth(vs sdktypes.Vars) bool {
	authType := vs.GetValueByString("auth_type")
	clientSecret := vs.GetValueByString("client_secret")
	return authType == integrations.OAuthPrivate || clientSecret != ""
}
