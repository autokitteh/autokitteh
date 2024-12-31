package sdktest

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// TestUsers is a test implementation of the Users service.
type TestUsers struct {
	Users map[sdktypes.UserID]sdktypes.User
}

func (t *TestUsers) Create(context.Context, sdktypes.User) (sdktypes.UserID, error) {
	return sdktypes.InvalidUserID, sdkerrors.ErrNotImplemented
}

func (t *TestUsers) Get(_ context.Context, id sdktypes.UserID, email string) (sdktypes.User, error) {
	if user, ok := t.Users[id]; ok {
		return user, nil
	}

	return sdktypes.InvalidUser, sdkerrors.ErrNotFound
}

func (t *TestUsers) Update(context.Context, sdktypes.User, *sdktypes.FieldMask) error {
	return sdkerrors.ErrNotImplemented
}
