package height

import (
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var apiKeyVar = sdktypes.NewSymbol("api_key")

// privateOAuth contains the user-provided details of a private Height OAuth 2.0 app.
type privateOAuth struct {
	ClientID     string `var:"private_client_id"`
	ClientSecret string `var:"private_client_secret,secret"`
}
