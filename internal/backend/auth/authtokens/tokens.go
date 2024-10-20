package authtokens

import "go.autokitteh.dev/autokitteh/sdk/sdktypes"

type Tokens interface {
	Create(userID sdktypes.UserID) (string, error)
	Parse(token string) (sdktypes.UserID, error)
}
