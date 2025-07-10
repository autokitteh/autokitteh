package authtokens

import "go.autokitteh.dev/autokitteh/sdk/sdktypes"

type Tokens interface {
	Create(user sdktypes.User) (string, error)
	CreateInternal(data map[string]string) (string, error)
	Parse(token string) (sdktypes.User, error)
	ParseInternal(token string) (map[string]string, error)
}
