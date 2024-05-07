package authtokens

import "go.autokitteh.dev/autokitteh/sdk/sdktypes"

type Tokens interface {
	Create(userID sdktypes.User) (string, error)
	Parse(token string) (sdktypes.User, error)
}
