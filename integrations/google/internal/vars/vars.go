package vars

import (
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Vars struct {
	OAuthData string `var:"secret"`
	JSON      string `var:"secret"`

	FormID               string
	FormResponsesWatchID string
	FormSchemaWatchID    string
}

var (
	OAuthData = sdktypes.NewSymbol("OAuthData")
	JSON      = sdktypes.NewSymbol("JSON")

	FormID               = sdktypes.NewSymbol("FormID")
	FormResponsesWatchID = sdktypes.NewSymbol("FormResponsesWatchID")
	FormSchemaWatchID    = sdktypes.NewSymbol("FormSchemaWatchID")
)
