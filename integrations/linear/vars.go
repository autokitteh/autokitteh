package linear

import (
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	actorVar  = sdktypes.NewSymbol("actor")
	apiKeyVar = sdktypes.NewSymbol("api_key")
)

// privateOAuth contains the user-provided details of a private Linear OAuth 2.0 app.
type privateOAuth struct {
	ClientID      string `var:"private_client_id"`
	ClientSecret  string `var:"private_client_secret,secret"`
	WebhookSecret string `var:"private_webhook_secret,secret"`
}
