package integrations

// Align these with:
// https://github.com/autokitteh/web-platform/blob/main/src/enums/connections/connectionTypes.enum.ts
const (
	APIKey       = "apiKey"
	APIToken     = "apiToken"
	Init         = "initialized" // integrations with 1 auth method
	JSONKey      = "jsonKey"     // Google integrations only
	OAuth        = "oauth"       // Deprecate?
	OAuthCustom  = "oauthCustom"
	OAuthDefault = "oauthDefault"
	PAT          = "pat"
	SocketMode   = "socketMode" // Slack integration only
)
