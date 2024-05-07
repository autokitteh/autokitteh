package oauth

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/chat/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/forms/v1"
	"google.golang.org/api/gmail/v1"
	googleoauth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/sheets/v4"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type oauth struct {
	logger *zap.Logger

	// Configs and opts store registration data together.
	// If we replace these in-memory maps with persistent
	// storage, we will merge them into a single table.
	configs map[string]*oauth2.Config
	opts    map[string]map[string]string
}

func New(l *zap.Logger) sdkservices.OAuth {
	// TODO(ENG-112): Remove (see Register below).
	redirectURL := fmt.Sprintf("https://%s/oauth/redirect/", os.Getenv("WEBHOOK_ADDRESS"))

	githubBaseURL := os.Getenv("GITHUB_ENTERPRISE_URL")
	if githubBaseURL == "" {
		githubBaseURL = "https://github.com"
	}
	var err error
	githubBaseURL, err = kittehs.NormalizeURL(githubBaseURL, true)
	if err != nil {
		l.Fatal("Invalid environment variable value",
			zap.String("name", "GITHUB_ENTERPRISE_URL"),
			zap.Error(err),
		)
	}

	appsDir := "apps"
	if os.Getenv("GITHUB_ENTERPRISE_URL") != "" {
		appsDir = "github-apps"
	}

	return &oauth{
		logger: l,
		// TODO(ENG-112): Construct the following 2 maps with dynamic integration
		// registrations, where each integration registration will call Register
		// below (if it uses OAuth). This hard-coding is EXTREMELY TEMPORARY!
		configs: map[string]*oauth2.Config{
			// Based on:
			// https://github.com/organizations/autokitteh/settings/apps/autokitteh
			"github": {
				ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
				ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
				Endpoint: oauth2.Endpoint{
					// https://docs.github.com/en/apps/using-github-apps/installing-a-github-app-from-a-third-party#installing-a-github-app
					AuthURL: fmt.Sprintf("%s/%s/%s/installations/new", githubBaseURL, appsDir, os.Getenv("GITHUB_APP_NAME")),
					// https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/authorizing-oauth-apps#device-flow
					DeviceAuthURL: fmt.Sprintf("%s/login/device/code", githubBaseURL),
					// https://docs.github.com/en/enterprise-server/apps/oauth-apps/building-oauth-apps/authorizing-oauth-apps#2-users-are-redirected-back-to-your-site-by-github
					// https://docs.github.com/en/enterprise-server/apps/sharing-github-apps/making-your-github-app-available-for-github-enterprise-server#the-app-code-must-use-the-correct-urls
					TokenURL: fmt.Sprintf("%s/login/oauth/access_token", githubBaseURL),
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
					// Sensitive.
					chat.ChatMembershipsScope,
					chat.ChatMessagesScope,
					chat.ChatSpacesScope,
					forms.FormsBodyScope,
					forms.FormsResponsesReadonlyScope,
					sheets.SpreadsheetsScope,
					// Restricted.
					drive.DriveScope,
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
					// Restricted.
					drive.DriveScope,
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
		opts: map[string]map[string]string{
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

func (o *oauth) StartFlow(ctx context.Context, intg string, cid sdktypes.ConnectionID) (string, error) {
	cfg, opts, err := o.Get(ctx, intg)
	if err != nil {
		return "", err
	}

	return cfg.AuthCodeURL(cid.String(), authCode(opts)...), nil
}

func (o *oauth) Exchange(ctx context.Context, intg, state, code string) (*oauth2.Token, error) {
	// Convert the received temporary authorization code
	// into a refresh token / user access token.
	cfg, opts, err := o.Get(ctx, intg)
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
