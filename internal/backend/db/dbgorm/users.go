package dbgorm

import (
	"context"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (gdb *gormdb) GetUserByID(ctx context.Context, id sdktypes.UserID) (sdktypes.User, error) {
	if id == sdktypes.DefaultUser.ID() {
		return sdktypes.DefaultUser, nil
	}

	var r scheme.User
	err := gdb.db.WithContext(ctx).Where("user_id = ?", id.UUIDValue()).First(&r).Error
	if err != nil {
		return sdktypes.InvalidUser, translateError(err)
	}

	return scheme.ParseUser(r), nil
}

func (gdb *gormdb) GetUserByEmail(ctx context.Context, email string) (sdktypes.User, error) {
	if email == sdktypes.DefaultUser.Email() {
		return sdktypes.DefaultUser, nil
	}

	var r scheme.User
	err := gdb.db.WithContext(ctx).Where("email = ?", email).First(&r).Error
	if err != nil {
		return sdktypes.InvalidUser, translateError(err)
	}

	return scheme.ParseUser(r), nil
}

func (gdb *gormdb) CreateUser(ctx context.Context, u sdktypes.User) (sdktypes.UserID, error) {
	if u.Email() == sdktypes.DefaultUser.Email() {
		return sdktypes.InvalidUserID, sdkerrors.NewInvalidArgumentError("cannot create default user")
	}

	uid := u.ID()
	if !uid.IsValid() {
		uid = sdktypes.NewUserID()
	}

	user := scheme.User{
		UserID:      uid.UUIDValue(),
		Email:       u.Email(),
		DisplayName: u.DisplayName(),
	}

	err := gdb.db.WithContext(ctx).Create(&user).Error
	if err != nil {
		return sdktypes.InvalidUserID, translateError(err)
	}

	return uid, nil
}
