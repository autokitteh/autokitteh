package authcontext

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type ctxKey string

const userIDCtxKey = ctxKey("user")

func GetAuthnUserID(ctx context.Context) sdktypes.UserID {
	if v := ctx.Value(userIDCtxKey); v != nil {
		return v.(sdktypes.UserID)
	}

	return sdktypes.InvalidUserID
}

func SetAuthnUserID(ctx context.Context, userID sdktypes.UserID) context.Context {
	if !userID.IsValid() {
		return ctx
	}

	return context.WithValue(ctx, userIDCtxKey, userID)
}
