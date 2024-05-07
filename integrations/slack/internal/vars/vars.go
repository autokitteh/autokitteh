package vars

import (
	"fmt"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Vars struct {
	AppID        string
	EnterpriseID string
	TeamID       string

	AppToken string `var:"secret"`
}

var (
	BotTokenName  = sdktypes.NewSymbol("bot_token")
	KeyName       = sdktypes.NewSymbol("key")
	WebSocketName = sdktypes.NewSymbol("websocket")
	OAuthDataName = sdktypes.NewSymbol("oauth_data")
)

func KeyValue(appID, enterpriseID, teamID string) string {
	return fmt.Sprintf("%s/%s/%s", appID, enterpriseID, teamID)
}
