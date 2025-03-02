package integrations

// Align these with:
// https://github.com/autokitteh/web-platform/blob/main/src/enums/connections/connectionTypes.enum.ts
const (
	APIKey         = "apiKey"
	APIToken       = "apiToken"
	DaemonApp      = "daemonApp"   // Microsoft integrations only
	Init           = "initialized" // Integrations with only 1 type by definition
	JSONKey        = "jsonKey"     // Google integrations only
	OAuth          = "oauth"       // Deprecate in the future, use "OAuthDefault"
	OAuthDefault   = "oauthDefault"
	OAuthPrivate   = "oauthPrivate"
	PAT            = "pat"
	ServerToServer = "serverToServer" // Zoom integrations only
	SocketMode     = "socketMode"     // Slack integration only
)
