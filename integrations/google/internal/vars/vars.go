package vars

import (
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Vars struct {
	OAuthData string `var:"secret"`
	JSON      string `var:"secret"`
}

var (
	OAuthData = sdktypes.NewSymbol("OAuthData")
	JSON      = sdktypes.NewSymbol("JSON")
)
