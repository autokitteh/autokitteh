package authcontext

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type ctxKey string

const userIDCtxKey = ctxKey("user")

func GetAuthnUser(ctx context.Context) sdktypes.User {
	if v := ctx.Value(userIDCtxKey); v != nil {
		return v.(sdktypes.User)
	}

	return sdktypes.InvalidUser
}

func SetAuthnUser(ctx context.Context, user sdktypes.User) context.Context {
	if !user.IsValid() {
		return ctx
	}

	return context.WithValue(ctx, userIDCtxKey, user)
}
