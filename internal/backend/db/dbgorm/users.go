package dbgorm

import (
	"context"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authusers"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (gdb *gormdb) GetUser(ctx context.Context, id sdktypes.UserID, name sdktypes.Symbol, email string) (sdktypes.User, error) {
	if id == authusers.DefaultUser.ID() {
		return authusers.DefaultUser, nil
	}

	q := gdb.db.WithContext(ctx)

	if !id.IsValid() && email == "" && !name.IsValid() {
		return sdktypes.InvalidUser, sdkerrors.NewInvalidArgumentError("missing id, email or name")
	}

	if id.IsValid() {
		q = q.Where("user_id = ?", id.UUIDValue())
	}

	if email != "" {
		q = q.Where("email = ?", email)
	}

	if name.IsValid() {
		q = q.Where("name = ?", name.String())
	}

	var r scheme.User
	err := q.First(&r).Error
	if err != nil {
		return sdktypes.InvalidUser, translateError(err)
	}

	return scheme.ParseUser(r)
}

func (gdb *gormdb) CreateUser(ctx context.Context, u sdktypes.User) (sdktypes.UserID, error) {
	uid := u.ID()
	if !uid.IsValid() {
		uid = sdktypes.NewUserID()
	} else if authusers.IsInternalUserID(uid) {
		return sdktypes.InvalidUserID, sdkerrors.ErrAlreadyExists
	}

	user := scheme.User{
		Base: based(ctx),

		UserID:       uid.UUIDValue(),
		Email:        u.Email(),
		DisplayName:  u.DisplayName(),
		Disabled:     u.Disabled(),
		Name:         u.Name().String(),
		DefaultOrgID: u.DefaultOrgID().UUIDValue(),
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

	if authusers.IsInternalUserID(u.ID()) {
		return sdkerrors.ErrUnauthorized
	}

	data := updatedBaseColumns(ctx)
	data["display_name"] = u.DisplayName()
	data["name"] = u.Name()
	data["disabled"] = u.Disabled()
	data["default_org_id"] = u.DefaultOrgID().UUIDValue()

	return translateError(
		gdb.db.WithContext(ctx).
			Model(&scheme.User{}).
			Where("user_id = ?", u.ID().UUIDValue()).
			Updates(data).
			Error,
	)
}
