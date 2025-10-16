package linear

import (
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	actorVar  = sdktypes.NewSymbol("actor")
	apiKeyVar = sdktypes.NewSymbol("api_key")
	orgIDVar  = sdktypes.NewSymbol("org_id")
)

const linearAPIURL = "https://api.linear.app/graphql"

// privateOAuth contains the user-provided details of a private Linear OAuth 2.0 app.
type privateOAuth struct {
	ClientID      string `var:"private_client_id"`
	ClientSecret  string `var:"private_client_secret,secret"`
	WebhookSecret string `var:"private_webhook_secret,secret"`
}

// orgInfo contains the details of a Linear organization being connected to.
type orgInfo struct {
	ID     string `json:"id" var:"org_id"`
	Name   string `json:"name" var:"org_name"`
	URLKey string `json:"urlKey" var:"org_url_key"`
}

// viewerInfo contains the user details of a Linear actor connected to an organization.
type viewerInfo struct {
	ID          string `json:"id" var:"viewer_id"`
	DisplayName string `json:"displayName" var:"viewer_display_name"`
	Email       string `json:"email" var:"viewer_email"`
	Name        string `json:"name" var:"viewer_name"`
}
