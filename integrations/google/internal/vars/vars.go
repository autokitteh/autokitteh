package vars

import (
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Vars struct {
	OAuthData string `var:"secret"`
	JSONKey   string `var:"secret"`
}

var (
	OAuthData = sdktypes.NewSymbol("OAuthData")
	JSONKey   = sdktypes.NewSymbol("JSONKey")
)
