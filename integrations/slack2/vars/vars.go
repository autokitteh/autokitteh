package vars

import (
	"fmt"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	SigningSecretEnvVar = "SLACK_SIGNING_SECRET"
)

var (
	// OAuth v2 app.
	SigningSecretVar = sdktypes.NewSymbol("private_signing_secret")

	// Socket Mode app.
	AppTokenVar = sdktypes.NewSymbol("private_app_token")
	BotTokenVar = sdktypes.NewSymbol("private_bot_token")

	// Install info.
	AppIDVar      = sdktypes.NewSymbol("app_id")
	InstallIDsVar = sdktypes.NewSymbol("install_ids")
)

// PrivateOAuth contains the user-provided details of a private OAuth v2 app.
type PrivateOAuth struct {
	ClientID      string `var:"private_client_id"`
	ClientSecret  string `var:"private_client_secret,secret"`
	SigningSecret string `var:"private_signing_secret,secret"`
}

// SocketMode contains the user-provided details of a private Socket Mode app.
type SocketMode struct {
	AppToken string `var:"private_app_token,secret"`
	BotToken string `var:"private_bot_token,secret"`
}

// InstallInfo contains the details of a Slack app's installation.
type InstallInfo struct {
	// Auth test.
	EnterpriseID string `var:"enterprise_id"`
	Team         string `var:"team_name"`
	TeamID       string `var:"team_id"`
	User         string `var:"user_name"`
	UserID       string `var:"user_id"`

	// Bot info.
	BotName    string `var:"bot_name"`
	BotID      string `var:"bot_id"`
	BotUpdated string `var:"bot_updated"`
	AppID      string `var:"app_id"`

	// For event routing.
	InstallIDs string `var:"install_ids"`
}

func InstallIDs(appID, enterpriseID, teamID string) string {
	return fmt.Sprintf("%s/%s/%s", appID, enterpriseID, teamID)
}
