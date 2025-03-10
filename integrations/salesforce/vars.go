package salesforce

import "go.autokitteh.dev/autokitteh/sdk/sdktypes"

var (
	clientIDVar    = sdktypes.NewSymbol("private_client_id")
	instanceURLVar = sdktypes.NewSymbol("instance_url")
	orgIDVar       = sdktypes.NewSymbol("organization_id")
)

// privateOAuth contains the user-provided details of a private Salesforce OAuth 2.0 app.
type privateOAuth struct {
	ClientID     string `var:"private_client_id"`
	ClientSecret string `var:"private_client_secret,secret"`
}
