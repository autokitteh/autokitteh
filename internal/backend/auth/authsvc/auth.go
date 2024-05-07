package authsvc

import (
	"context"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authtokens"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type auth struct {
	tokens authtokens.Tokens
}

func New(tokens authtokens.Tokens) sdkservices.Auth { return &auth{tokens: tokens} }

func (a *auth) WhoAmI(ctx context.Context) (sdktypes.User, error) {
	user := authcontext.GetAuthnUser(ctx)
	if !user.IsValid() {
		return sdktypes.InvalidUser, nil
	}

	return user, nil
}

func (a *auth) CreateToken(ctx context.Context) (string, error) {
	user := authcontext.GetAuthnUser(ctx)
	if !user.IsValid() {
		return "", sdkerrors.ErrUnauthenticated
	}

	if a.tokens == nil {
		return "", sdkerrors.ErrNotImplemented
	}

	return a.tokens.Create(user)
}
