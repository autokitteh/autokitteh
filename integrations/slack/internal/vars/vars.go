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
	// Socket Mode
	AppTokenName = sdktypes.NewSymbol("AppToken")
	BotTokenName = sdktypes.NewSymbol("BotToken")

	KeyName       = sdktypes.NewSymbol("Key")
	OAuthDataName = sdktypes.NewSymbol("OAuthData")
)

func KeyValue(appID, enterpriseID, teamID string) string {
	return fmt.Sprintf("%s/%s/%s", appID, enterpriseID, teamID)
}
