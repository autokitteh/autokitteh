package reddit

import (
	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var desc = common.Descriptor("reddit", "Reddit", "/static/images/reddit.svg")

var redditTokenURL = "https://www.reddit.com/api/v1/access_token"

var (
	clientIDVar     = sdktypes.NewSymbol("client_id")
	clientSecretVar = sdktypes.NewSymbol("client_secret")
	userAgentVar    = sdktypes.NewSymbol("user_agent")
	usernameVar     = sdktypes.NewSymbol("username")
	passwordVar     = sdktypes.NewSymbol("password")
)
