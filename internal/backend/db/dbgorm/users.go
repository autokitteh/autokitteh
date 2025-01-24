package dbgorm

import (
	"context"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authusers"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (gdb *gormdb) GetUser(ctx context.Context, id sdktypes.UserID, email string) (sdktypes.User, error) {
	q := gdb.rdb.WithContext(ctx)

	if !id.IsValid() && email == "" {
		return sdktypes.InvalidUser, sdkerrors.NewInvalidArgumentError("missing id or email")
	}

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

	return scheme.ParseUser(r)
}

func (gdb *gormdb) CreateUser(ctx context.Context, u sdktypes.User) (sdktypes.UserID, error) {
	if u.Status() == sdktypes.UserStatusUnspecified {
		return sdktypes.InvalidUserID, sdkerrors.NewInvalidArgumentError("missing status")
	}

	uid := u.ID()
	if !uid.IsValid() {
		uid = sdktypes.NewUserID()
	} else if authusers.IsSystemUserID(uid) {
		return sdktypes.InvalidUserID, sdkerrors.NewInvalidArgumentError("system user")
	}

	user := scheme.User{
		Base: based(ctx),

		UserID:       uid.UUIDValue(),
		Email:        u.Email(),
		DisplayName:  u.DisplayName(),
		DefaultOrgID: u.DefaultOrgID().UUIDValue(),
		Status:       int32(u.Status().ToProto()),
	}

	err := gdb.wdb.WithContext(ctx).Create(&user).Error
	if err != nil {
		return sdktypes.InvalidUserID, translateError(err)
	}

	return uid, nil
}

func (gdb *gormdb) UpdateUser(ctx context.Context, u sdktypes.User, fm *sdktypes.FieldMask) error {
	if !u.ID().IsValid() {
		return sdkerrors.NewInvalidArgumentError("missing uid")
	}

	if authusers.IsSystemUserID(u.ID()) {
		return sdkerrors.ErrUnauthorized
	}

	data, err := updatedFields(ctx, u, fm)
	if err != nil {
		return err
	}

	return translateError(
		gdb.wdb.WithContext(ctx).
			Model(&scheme.User{}).
			Where("user_id = ?", u.ID().UUIDValue()).
			Updates(data).
			Error,
	)
}
