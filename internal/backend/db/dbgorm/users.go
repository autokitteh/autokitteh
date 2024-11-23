package dbgorm

import (
	"context"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authusers"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (gdb *gormdb) GetUser(ctx context.Context, id sdktypes.UserID, email string) (sdktypes.User, error) {
	if id == authusers.DefaultUser.ID() {
		return authusers.DefaultUser, nil
	}

	q := gdb.db.WithContext(ctx)

	if id.IsValid() {
		q = q.Where("user_id = ?", id.UUIDValue())
	}

	if email != "" {
		q = q.Where("email = ?", email)
	}

	var r scheme.User
	err := q.First(&r).Error
	if err != nil {
		return sdktypes.InvalidUser, translateError(err)
	}

	return scheme.ParseUser(r), nil
}

func (gdb *gormdb) CreateUser(ctx context.Context, u sdktypes.User) (sdktypes.UserID, error) {
	uid := u.ID()
	if !uid.IsValid() {
		uid = sdktypes.NewUserID()
	} else if authusers.IsInternalUserID(uid) {
		return sdktypes.InvalidUserID, sdkerrors.ErrAlreadyExists
	}

	user := scheme.User{
		UserID:      uid.UUIDValue(),
		Email:       u.Email(),
		DisplayName: u.DisplayName(),
		Disabled:    u.Disabled(),
	}

	err := gdb.db.WithContext(ctx).Create(&user).Error
	if err != nil {
		return sdktypes.InvalidUserID, translateError(err)
	}

	return uid, nil
}

func (gdb *gormdb) UpdateUser(ctx context.Context, u sdktypes.User) error {
	if !u.ID().IsValid() {
		return sdkerrors.NewInvalidArgumentError("missing uid")
	}

	return translateError(
		gdb.db.WithContext(ctx).
			Model(&scheme.User{}).
			Update("display_name", u.DisplayName()).Update("disabled", u.Disabled()).
			Where("user_id = ?", u.ID().UUIDValue()).
			Error,
	)
}
