package authcontext

import (
	"context"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authusers"

	"go.autokitteh.dev/autokitteh/sdk/sdklogger"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type ctxKey string

var userCtxKey = ctxKey("user")

// Get the authenticated user.
func GetAuthnUser(ctx context.Context) sdktypes.User {
	if v := ctx.Value(userCtxKey); v != nil {
		return v.(sdktypes.User)
	}
	return sdktypes.InvalidUser
}

// This function is used to set the authenticated user in the context.
func SetAuthnUser(ctx context.Context, user sdktypes.User) context.Context {
	if !user.IsValid() {
		return ctx
	}
	return context.WithValue(ctx, userCtxKey, user)
}

// Helper function.
func GetAuthnUserID(ctx context.Context) sdktypes.UserID {
	return GetAuthnUser(ctx).ID()
}

// Use this function to fill default org id for queries.
// In such queries we do not need to query for system users, but all users.
// If the user is not authenticated, it will return InvalidOrgID.
// If the user is a system user, it will return InvalidOrgID.
// Otherwise, will return the user's default org id.
func GetAuthnInferredOrgID(ctx context.Context) sdktypes.OrgID {
	if IsAuthnSystemUser(ctx) {
		return sdktypes.InvalidOrgID
	}

	return GetAuthnUser(ctx).DefaultOrgID()
}

func SetAuthnSystemUser(ctx context.Context) context.Context {
	return SetAuthnUser(ctx, authusers.SystemUser)
}

func IsAuthnSystemUser(ctx context.Context) bool {
	return GetAuthnUser(ctx).ID() == authusers.SystemUser.ID()
}

// If `o` has an owner, return unmodified.
// Otherwise, set the owner to the authenticated user.
func ObjectWithOrgID[O interface {
	OrgID() sdktypes.OrgID
	WithOrgID(sdktypes.OrgID) O
}](ctx context.Context, o O) O {
	if o.OrgID().IsValid() {
		return o
	}

	oid := GetAuthnInferredOrgID(ctx)

	if !oid.IsValid() {
		sdklogger.DPanic("no org id")
	}

	return o.WithOrgID(oid)
}
