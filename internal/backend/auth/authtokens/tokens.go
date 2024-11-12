package authtokens

import "go.autokitteh.dev/autokitteh/sdk/sdktypes"

type Tokens interface {
	Create(user sdktypes.User) (string, error)
	Parse(token string) (sdktypes.User, error)
}
