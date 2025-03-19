package authsvc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type tokens struct{}

func (tokens) Create(user sdktypes.User) (string, error) { return user.ID().String(), nil }
func (tokens) Parse(token string) (sdktypes.User, error) {
	return sdktypes.InvalidUser, sdkerrors.ErrNotImplemented
}

func TestCreateToken(t *testing.T) {
	u := sdktypes.NewUser().WithNewID()

	tok, err := New(nil).CreateToken(authcontext.SetAuthnUser(t.Context(), u))
	assert.ErrorIs(t, err, sdkerrors.ErrNotImplemented)
	assert.Equal(t, tok, "")

	a := New(tokens{})

	tok, err = a.CreateToken(context.Background())
	assert.ErrorIs(t, err, sdkerrors.ErrUnauthenticated)
	assert.Equal(t, tok, "")

	tok, err = a.CreateToken(authcontext.SetAuthnSystemUser(t.Context()))
	assert.ErrorIs(t, err, sdkerrors.ErrUnauthorized)
	assert.Equal(t, tok, "")

	tok, err = a.CreateToken(authcontext.SetAuthnUser(t.Context(), u))
	if assert.NoError(t, err) {
		assert.Equal(t, tok, u.ID().String())
	}
}
