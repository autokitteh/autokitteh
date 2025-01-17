package dbgorm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authusers"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var defaultUserMembership = sdktypes.NewOrgMember(authusers.DefaultOrg.ID(), authusers.DefaultUser.ID()).
	WithStatus(sdktypes.OrgMemberStatusActive).
	WithRoles(sdktypes.NewSymbol("admin"))

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

func (gdb *gormdb) BatchGetOrgs(ctx context.Context, oids []sdktypes.OrgID) ([]sdktypes.Org, error) {
	if len(oids) > maxBatchSize {
		return nil, sdkerrors.NewInvalidArgumentError("too many orgs = %d > %d", len(oids), maxBatchSize)
	}

	q := gdb.db.WithContext(ctx).Where("org_id in ?", kittehs.Transform(oids, func(oid sdktypes.OrgID) uuid.UUID { return oid.UUIDValue() }))

	var rs []scheme.Org
	err := q.Find(&rs).Error
	if err != nil {
		return nil, translateError(err)
	}

	orgs, err := kittehs.TransformError(rs, scheme.ParseOrg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse orgs: %w", err)
	}

	if kittehs.ContainedIn(oids...)(authusers.DefaultOrg.ID()) {
		orgs = append(orgs, authusers.DefaultOrg)
	}

	return orgs, err
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
		Name:        o.Name().String(),
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

func (gdb *gormdb) ListOrgMembers(ctx context.Context, oid sdktypes.OrgID) ([]sdktypes.OrgMember, error) {
	if oid == authusers.DefaultOrg.ID() {
		return []sdktypes.OrgMember{defaultUserMembership}, nil
	}

	var ms []scheme.OrgMember
	err := gdb.db.WithContext(ctx).
		Where("org_id = ?", oid.UUIDValue()).
		Order("created_at ASC").Find(&ms).
		Error
	if err != nil {
		return nil, translateError(err)
	}

	return kittehs.TransformError(ms, scheme.ParseOrgMember)
}

func (gdb *gormdb) AddOrgMember(ctx context.Context, m sdktypes.OrgMember) error {
	if m.OrgID() == authusers.DefaultOrg.ID() {
		return sdkerrors.ErrUnauthorized
	}

	roles, err := json.Marshal(m.Roles())
	if err != nil {
		return fmt.Errorf("failed to marshal roles: %w", err)
	}

	ou := scheme.OrgMember{
		OrgID:  m.OrgID().UUIDValue(),
		UserID: m.UserID().UUIDValue(),
		Status: int(m.Status().ToProto()),
		Roles:  roles,
	}

	return translateError(gdb.db.WithContext(ctx).Create(&ou).Error)
}

func (gdb *gormdb) UpdateOrgMember(ctx context.Context, m sdktypes.OrgMember, fm *sdktypes.FieldMask) error {
	if m.OrgID() == authusers.DefaultOrg.ID() {
		return sdkerrors.ErrUnauthorized
	}

	data, err := updatedFields(ctx, m, fm)
	if err != nil {
		return err
	}

	if roles, ok := data["roles"]; ok {
		if data["roles"], err = json.Marshal(roles); err != nil {
			return fmt.Errorf("failed to marshal roles: %w", err)
		}
	}

	return translateError(
		gdb.db.WithContext(ctx).
			Model(&scheme.OrgMember{}).
			Where("org_id = ? AND user_id = ?", m.OrgID().UUIDValue(), m.UserID().UUIDValue()).
			Updates(data).
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

func (gdb *gormdb) GetOrgMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) (sdktypes.OrgMember, error) {
	if oid == authusers.DefaultOrg.ID() {
		if uid == authusers.DefaultUser.ID() {
			return defaultUserMembership, nil
		}

		return sdktypes.InvalidOrgMember, sdkerrors.ErrNotFound
	}

	var om scheme.OrgMember

	err := gdb.db.WithContext(ctx).
		Model(&scheme.OrgMember{}).
		Where("org_id = ? AND user_id = ?", oid.UUIDValue(), uid.UUIDValue()).
		First(&om).
		Error
	if err != nil {
		return sdktypes.InvalidOrgMember, translateError(err)
	}

	// Backward compatibility from prior to having a status.
	if om.Status == 0 {
		om.Status = int(sdktypes.OrgMemberStatusActive.ToProto())
	}

	return scheme.ParseOrgMember(om)
}

func (gdb *gormdb) GetOrgsForUser(ctx context.Context, uid sdktypes.UserID) ([]sdktypes.OrgMember, error) {
	if uid == authusers.DefaultUser.ID() {
		return []sdktypes.OrgMember{defaultUserMembership}, nil
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

	return kittehs.TransformError(oms, scheme.ParseOrgMember)
}
