package vars

import (
	"fmt"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Vars struct {
	AppID        string
	EnterpriseID string
	TeamID       string
}

var (
	AuthType = sdktypes.NewSymbol("auth_type")

	// Custom OAuth
	ClientID      = sdktypes.NewSymbol("client_id")
	ClientSecret  = sdktypes.NewSymbol("client_secret")
	SigningSecret = sdktypes.NewSymbol("signing_secret")

	// Socket Mode
	AppTokenName = sdktypes.NewSymbol("AppToken")
	BotTokenName = sdktypes.NewSymbol("BotToken")

	KeyName       = sdktypes.NewSymbol("Key")
	OAuthDataName = sdktypes.NewSymbol("OAuthData")
)

func KeyValue(appID, enterpriseID, teamID string) string {
	return fmt.Sprintf("%s/%s/%s", appID, enterpriseID, teamID)
}
