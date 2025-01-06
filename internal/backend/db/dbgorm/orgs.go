package dbgorm

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authusers"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (gdb *gormdb) GetOrg(ctx context.Context, oid sdktypes.OrgID) (sdktypes.Org, error) {
	if oid == authusers.DefaultOrg.ID() {
		return authusers.DefaultOrg, nil
	}

	q := gdb.db.WithContext(ctx)

	if !oid.IsValid() {
		return sdktypes.InvalidOrg, sdkerrors.NewInvalidArgumentError("missing id")
	}

	if oid.IsValid() {
		q = q.Where("org_id = ?", oid.UUIDValue())
	}

	var r scheme.Org
	err := q.First(&r).Error
	if err != nil {
		return sdktypes.InvalidOrg, translateError(err)
	}

	return scheme.ParseOrg(r)
}

func (gdb *gormdb) DeleteOrg(ctx context.Context, oid sdktypes.OrgID) error {
	if oid == authusers.DefaultOrg.ID() {
		return sdkerrors.ErrUnauthorized
	}

	if !oid.IsValid() {
		return sdkerrors.NewInvalidArgumentError("missing id")
	}

	return translateError(gdb.transaction(ctx, func(tx *tx) error {
		err := tx.db.Where("org_id = ?", oid.UUIDValue()).Delete(&scheme.OrgMember{}).Error
		if err != nil {
			return translateError(err)
		}

		return tx.db.Where("org_id = ?", oid.UUIDValue()).Delete(&scheme.Org{}).Error
	}))
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
		DisplayName: o.DisplayName(),
	}

	err := gdb.db.WithContext(ctx).Create(&org).Error
	if err != nil {
		return sdktypes.InvalidOrgID, translateError(err)
	}

	return oid, nil
}

func (gdb *gormdb) UpdateOrg(ctx context.Context, o sdktypes.Org, fm *sdktypes.FieldMask) error {
	if o.ID() == authusers.DefaultOrg.ID() {
		return sdkerrors.ErrUnauthorized
	}

	data, err := updatedFields(ctx, o, fm)
	if err != nil {
		return err
	}

	return translateError(
		gdb.db.WithContext(ctx).
			Model(&scheme.Org{}).
			Where("org_id = ?", o.ID().UUIDValue()).
			Updates(data).
			Error,
	)
}

func (gdb *gormdb) ListOrgMembers(ctx context.Context, oid sdktypes.OrgID) ([]*sdkservices.UserIDWithMemberStatus, error) {
	if oid == authusers.DefaultOrg.ID() {
		return []*sdkservices.UserIDWithMemberStatus{
			{
				UserID: authusers.DefaultUser.ID(),
				Status: sdktypes.OrgMemberStatusActive,
			},
		}, nil
	}

	var ous []scheme.OrgMember
	err := gdb.db.WithContext(ctx).
		Where("org_id = ?", oid.UUIDValue()).
		Order("created_at ASC").Find(&ous).
		Error
	if err != nil {
		return nil, translateError(err)
	}

	return kittehs.TransformError(ous, func(ou scheme.OrgMember) (*sdkservices.UserIDWithMemberStatus, error) {
		s, err := sdktypes.OrgMemberStatusFromProto(sdktypes.OrgMemberStatusPB(ou.Status))
		if err != nil {
			return nil, fmt.Errorf("failed to parse org member status %q: %w", ou.Status, err)
		}

		return &sdkservices.UserIDWithMemberStatus{
			UserID: sdktypes.NewIDFromUUID[sdktypes.UserID](ou.UserID),
			Status: s,
		}, nil
	})
}

func (gdb *gormdb) AddOrgMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID, status sdktypes.OrgMemberStatus) error {
	if oid == authusers.DefaultOrg.ID() {
		return sdkerrors.ErrUnauthorized
	}

	ou := scheme.OrgMember{
		OrgID:  oid.UUIDValue(),
		UserID: uid.UUIDValue(),
		Status: int(status.ToProto()),
	}

	return translateError(gdb.db.WithContext(ctx).Create(&ou).Error)
}

func (gdb *gormdb) UpdateOrgMemberStatus(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID, s sdktypes.OrgMemberStatus) error {
	if oid == authusers.DefaultOrg.ID() {
		return sdkerrors.ErrUnauthorized
	}

	return translateError(
		gdb.db.WithContext(ctx).
			Model(&scheme.OrgMember{}).
			Where("org_id = ? AND user_id = ?", oid.UUIDValue(), uid.UUIDValue()).
			Update("status", int(s.ToProto())).
			Error,
	)
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

func (gdb *gormdb) GetOrgMemberStatus(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) (sdktypes.OrgMemberStatus, error) {
	if oid == authusers.DefaultOrg.ID() {
		if uid == authusers.DefaultUser.ID() {
			return sdktypes.OrgMemberStatusActive, nil
		}

		return sdktypes.OrgMemberStatusUnspecified, nil
	}

	var om scheme.OrgMember

	err := gdb.db.WithContext(ctx).
		Model(&scheme.OrgMember{}).
		Where("org_id = ? AND user_id = ?", oid.UUIDValue(), uid.UUIDValue()).
		Select("status").
		First(&om).
		Error
	if err != nil {
		return sdktypes.OrgMemberStatusUnspecified, translateError(err)
	}

	// Backward compatibility from prior to having a status.
	if om.Status == 0 {
		om.Status = int(sdktypes.OrgMemberStatusActive.ToProto())
	}

	return sdktypes.OrgMemberStatusFromProto(sdktypes.OrgMemberStatusPB(om.Status))
}

func (gdb *gormdb) GetOrgsForUser(ctx context.Context, uid sdktypes.UserID) ([]*sdkservices.OrgWithMemberStatus, error) {
	if uid == authusers.DefaultUser.ID() {
		return []*sdkservices.OrgWithMemberStatus{
			{
				Org:    authusers.DefaultOrg,
				Status: sdktypes.OrgMemberStatusActive,
			},
		}, nil
	}

	var oms []scheme.OrgMember

	err := gdb.db.WithContext(ctx).
		Where("user_id = ?", uid.UUIDValue()).
		Order("created_at ASC").
		Find(&oms).
		Error
	if err != nil {
		return nil, translateError(err)
	}

	oids := kittehs.Transform(oms, func(om scheme.OrgMember) uuid.UUID { return om.OrgID })

	var orgs []scheme.Org

	err = gdb.db.WithContext(ctx).
		Where("org_id IN ?", oids).
		Find(&orgs).
		Error
	if err != nil {
		return nil, translateError(err)
	}

	orgsMap := make(map[uuid.UUID]sdktypes.Org, len(orgs))
	for _, o := range orgs {
		if orgsMap[o.OrgID], err = scheme.ParseOrg(o); err != nil {
			return nil, fmt.Errorf("failed to parse org %q: %w", o.OrgID, err)
		}
	}

	ows := make([]*sdkservices.OrgWithMemberStatus, len(oms))
	for i, om := range oms {
		s, err := sdktypes.OrgMemberStatusFromProto(sdktypes.OrgMemberStatusPB(om.Status))
		if err != nil {
			return nil, fmt.Errorf("failed to parse org member status %q: %w", om.Status, err)
		}

		ows[i] = &sdkservices.OrgWithMemberStatus{Org: orgsMap[om.OrgID], Status: s}
	}

	return ows, nil
}
