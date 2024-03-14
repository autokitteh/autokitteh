package authsvc

import (
	"context"

	"go.autokitteh.dev/autokitteh/ee/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/ee/internal/backend/auth/authtokens"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type auth struct {
	users  sdkservices.Users
	tokens authtokens.Tokens
}

func New(users sdkservices.Users, tokens authtokens.Tokens) sdkservices.Auth {
	return &auth{users: users, tokens: tokens}
}

func (a *auth) WhoAmI(ctx context.Context) (sdktypes.User, error) {
	userID := authcontext.GetAuthnUserID(ctx)
	if !userID.IsValid() {
		return sdktypes.InvalidUser, nil
	}

	return a.users.GetByID(ctx, userID)
}

func (a *auth) CreateToken(ctx context.Context) (string, error) {
	userID := authcontext.GetAuthnUserID(ctx)
	if !userID.IsValid() {
		return "", sdkerrors.ErrUnauthenticated
	}

	if a.tokens == nil {
		return "", sdkerrors.ErrNotImplemented
	}

	return a.tokens.Create(userID)
}
