package sdktest

import (
	"context"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// TestUsers is a test implementation of the Users service.
type TestUsers struct {
	Users map[sdktypes.UserID]sdktypes.User

	IDGen func() sdktypes.UserID

	CreateCalledCount, GetCalledCount, UpdateCalledCount int
}

func (t *TestUsers) Reset(users ...sdktypes.User) {
	t.Users = make(map[sdktypes.UserID]sdktypes.User)
	for _, u := range users {
		t.Users[u.ID()] = u
	}

	t.CreateCalledCount = 0
	t.GetCalledCount = 0
	t.UpdateCalledCount = 0
}

func (t *TestUsers) idgen() sdktypes.UserID {
	if t.IDGen != nil {
		return t.IDGen()
	}

	return sdktypes.NewUserID()
}

// CAUTION: This does no validations.
func (t *TestUsers) Create(_ context.Context, u sdktypes.User) (sdktypes.UserID, error) {
	t.CreateCalledCount++

	if t.Users == nil {
		t.Users = make(map[sdktypes.UserID]sdktypes.User)
	}

	u = u.WithID(t.idgen())
	t.Users[u.ID()] = u
	return u.ID(), nil
}

func (t *TestUsers) Get(_ context.Context, id sdktypes.UserID, email string) (sdktypes.User, error) {
	t.GetCalledCount++

	if id.IsValid() && email != "" {
		return sdktypes.InvalidUser, sdkerrors.NewInvalidArgumentError("id and email are mutually exclusive")
	}

	if !id.IsValid() && email == "" {
		return sdktypes.InvalidUser, sdkerrors.NewInvalidArgumentError("id or email must be specified")
	}

	if id.IsValid() {
		if user, ok := t.Users[id]; ok {
			return user, nil
		}
	} else {
		for _, user := range t.Users {
			if user.Email() == email {
				return user, nil
			}
		}
	}

	return sdktypes.InvalidUser, sdkerrors.ErrNotFound
}

// CAUTION: This does no validations.
func (t *TestUsers) Update(_ context.Context, u sdktypes.User, fm *sdktypes.FieldMask) error {
	t.UpdateCalledCount++

	uu, ok := t.Users[u.ID()]
	if !ok {
		return sdkerrors.ErrNotFound
	}

	if fm == nil {
		t.Users[u.ID()] = u
		return nil
	}

	has := kittehs.ContainedIn(fm.Paths...)

	if has("display_name") {
		uu = uu.WithDisplayName(u.DisplayName())
	}

	if has("status") {
		uu = uu.WithStatus(u.Status())
	}

	if has("default_org_id") {
		uu = uu.WithDefaultOrgID(u.DefaultOrgID())
	}

	t.Users[u.ID()] = uu

	return nil
}
