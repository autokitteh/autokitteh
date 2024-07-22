package authcontext

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type ctxKey string

var (
	userCtxKey   = ctxKey("user")
	userIDCtxKey = ctxKey("userid")
)

func GetAuthnUser(ctx context.Context) sdktypes.User {
	if v := ctx.Value(userCtxKey); v != nil {
		return v.(sdktypes.User)
	}
	return sdktypes.InvalidUser
}

func SetAuthnUser(ctx context.Context, user sdktypes.User) context.Context {
	if !user.IsValid() {
		return ctx
	}
	return context.WithValue(ctx, userCtxKey, user)
}

func GetAuthnUserID(ctx context.Context) string {
	if v := ctx.Value(userIDCtxKey); v != nil {
		return v.(string)
	}
	return ""
}

func SetAuthnUserID(ctx context.Context, userID string) context.Context {
	if userID == "" {
		return ctx
	}
	return context.WithValue(ctx, userIDCtxKey, userID)
}
