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
	AuthType = sdktypes.NewSymbol("authType")

	// Custom OAuth
	ClientID         = sdktypes.NewSymbol("clientID")
	ClientSecret     = sdktypes.NewSymbol("clientSecret")
	SigningSecret    = sdktypes.NewSymbol("signingSecret")

	// Socket Mode
	AppTokenName = sdktypes.NewSymbol("AppToken")
	BotTokenName = sdktypes.NewSymbol("BotToken")

	KeyName       = sdktypes.NewSymbol("Key")
	OAuthDataName = sdktypes.NewSymbol("OAuthData")
)

func KeyValue(appID, enterpriseID, teamID string) string {
	return fmt.Sprintf("%s/%s/%s", appID, enterpriseID, teamID)
}
