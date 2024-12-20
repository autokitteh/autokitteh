package dbgorm

import (
	"context"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authusers"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (gdb *gormdb) GetOrg(ctx context.Context, oid sdktypes.OrgID, n sdktypes.Symbol) (sdktypes.Org, error) {
	if oid == authusers.DefaultOrg.ID() {
		return authusers.DefaultOrg, nil
	}

	q := gdb.db.WithContext(ctx)

	if !oid.IsValid() && !n.IsValid() {
		return sdktypes.InvalidOrg, sdkerrors.NewInvalidArgumentError("missing id or name")
	}

	if oid.IsValid() {
		q = q.Where("org_id = ?", oid.UUIDValue())
	}

	if n.IsValid() {
		q = q.Where("name = ?", n.String())
	}

	var r scheme.Org
	err := q.First(&r).Error
	if err != nil {
		return sdktypes.InvalidOrg, translateError(err)
	}

	return scheme.ParseOrg(r)
}

func (gdb *gormdb) CreateOrg(ctx context.Context, o sdktypes.Org) (sdktypes.OrgID, error) {
	oid := o.ID()

	if oid == authusers.DefaultOrg.ID() {
		return sdktypes.InvalidOrgID, sdkerrors.ErrAlreadyExists
	}

	if !oid.IsValid() {
		oid = sdktypes.NewOrgID()
	}

	org := scheme.Org{
		Base: based(ctx),

		OrgID:       oid.UUIDValue(),
		Name:        o.Name().String(),
		DisplayName: o.DisplayName(),
	}

	err := gdb.db.WithContext(ctx).Create(&org).Error
	if err != nil {
		return sdktypes.InvalidOrgID, translateError(err)
	}

	return oid, nil
}

func (gdb *gormdb) UpdateOrg(ctx context.Context, o sdktypes.Org) error {
	if o.ID() == authusers.DefaultOrg.ID() {
		return sdkerrors.ErrUnauthorized
	}

	data := updatedBaseColumns(ctx)
	data["name"] = o.Name().String()
	data["display_name"] = o.DisplayName()

	return translateError(
		gdb.db.WithContext(ctx).
			Model(&scheme.Org{}).
			Updates(data).
			Where("org_id = ?", o.ID().UUIDValue()).
			Error,
	)
}

func (gdb *gormdb) ListOrgMembers(ctx context.Context, oid sdktypes.OrgID) ([]sdktypes.UserID, error) {
	if oid == authusers.DefaultOrg.ID() {
		return []sdktypes.UserID{authusers.DefaultUser.ID()}, nil
	}

	var ous []scheme.OrgMember
	if err := gdb.db.WithContext(ctx).Where("org_id = ?", oid.UUIDValue()).Find(&ous).Error; err != nil {
		return nil, translateError(err)
	}

	return kittehs.Transform(ous, func(ou scheme.OrgMember) sdktypes.UserID {
		return sdktypes.NewIDFromUUID[sdktypes.UserID](ou.UserID)
	}), nil
}

func (gdb *gormdb) AddOrgMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) error {
	if oid == authusers.DefaultOrg.ID() {
		return sdkerrors.ErrUnauthorized
	}

	ou := scheme.OrgMember{
		OrgID:  oid.UUIDValue(),
		UserID: uid.UUIDValue(),
	}

	return translateError(gdb.db.WithContext(ctx).Create(&ou).Error)
}

func (gdb *gormdb) RemoveOrgMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) error {
	if oid == authusers.DefaultOrg.ID() {
		return sdkerrors.ErrUnauthorized
	}

	return translateError(
		gdb.db.WithContext(ctx).
			Where("org_id = ? AND user_id = ?", oid.UUIDValue(), uid.UUIDValue()).
			Delete(&scheme.OrgMember{}).
			Error,
	)
}

func (gdb *gormdb) IsOrgMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) (bool, error) {
	if oid == authusers.DefaultOrg.ID() {
		return uid == authusers.DefaultUser.ID(), nil
	}

	var count int64
	err := gdb.db.WithContext(ctx).
		Model(&scheme.OrgMember{}).
		Where("org_id = ? AND user_id = ?", oid.UUIDValue(), uid.UUIDValue()).
		Count(&count).
		Error
	if err != nil {
		return false, translateError(err)
	}

	return count > 0, nil
}

func (gdb *gormdb) GetOrgsForUser(ctx context.Context, uid sdktypes.UserID) ([]sdktypes.OrgID, error) {
	if uid == authusers.DefaultUser.ID() {
		return []sdktypes.OrgID{authusers.DefaultOrg.ID()}, nil
	}

	var ous []scheme.OrgMember

	err := gdb.db.WithContext(ctx).
		Model(&scheme.OrgMember{}).
		Where("user_id = ?", uid.UUIDValue()).
		Find(&ous).
		Error
	if err != nil {
		return nil, translateError(err)
	}

	return kittehs.Transform(ous, func(ou scheme.OrgMember) sdktypes.OrgID {
		return sdktypes.NewIDFromUUID[sdktypes.OrgID](ou.OrgID)
	}), nil
}
