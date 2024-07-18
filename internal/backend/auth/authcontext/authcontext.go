package authcontext

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type ctxKey string

var (
	userCtxKey              = ctxKey("user")
	userIDCtxKey            = ctxKey("userid")
	ownershipRequiredCtxKey = ctxKey("ownership")
	componentCtxKey         = ctxKey("component")
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

func NeedOwnershipCheck(ctx context.Context) (checkOwnership bool) {
	if v := ctx.Value(ownershipRequiredCtxKey); v != nil {
		checkOwnership = v.(bool)
	}
	if !checkOwnership { // require ownerShip check if user is in context as well
		checkOwnership = ctx.Value(userCtxKey) != nil
	}
	return
}

func SetOwnershipRequired(ctx context.Context) context.Context {
	return context.WithValue(ctx, ownershipRequiredCtxKey, true)
}

func GetAuthnUserID(ctx context.Context) string {
	if v := ctx.Value(userIDCtxKey); v != nil {
		return v.(string)
	}
	return ""
}

func SetAuthnUserID(ctx context.Context, uid string) context.Context {
	return context.WithValue(ctx, userIDCtxKey, uid)
}

func GetComponent(ctx context.Context) string {
	if v := ctx.Value(componentCtxKey); v != nil {
		return v.(string)
	}

	return "unknown"
}

func SetComponent(ctx context.Context, initiator string) context.Context {
	return context.WithValue(ctx, componentCtxKey, initiator)
}

// func WithUserCtx(ctx context.Context, db db.DB, id sdktypes.UUID) context.Context {
// 	uid, err := db.GetOwnership(ctx, id)
// 	fmt.Println("WITH USER CTX ID uid", uid, err)
// 	// FIXME: log!
// 	// if err != nil {
// 	// return ctx, fmt.Errorf("get ownership: %w", err)
// 	// }

// 	c := SetAuthnUserID(ctx, uid)
// 	u := GetAuthnUser(c)
// 	id1 := GetAuthnUserID(c)
// 	fmt.Println("WITH USER CTX test", u, id1)
// 	return c
// }
