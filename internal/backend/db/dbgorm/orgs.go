package dbgorm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (gdb *gormdb) GetOrg(ctx context.Context, oid sdktypes.OrgID, n sdktypes.Symbol) (sdktypes.Org, error) {
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

func (gdb *gormdb) DeleteOrg(ctx context.Context, oid sdktypes.OrgID) error {
	if !oid.IsValid() {
		return sdkerrors.NewInvalidArgumentError("missing id")
	}

	projectCount, err := gdb.CountProjects(ctx, oid)
	if err != nil {
		return translateError(err)
	}
	if projectCount > 0 {
		return errors.New("cannot delete an organization with projects")
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

	if !oid.IsValid() {
		oid = sdktypes.NewOrgID()
	}

	err := gdb.transaction(ctx, func(tx *tx) error {
		org := scheme.Org{
			Base: based(ctx),

			OrgID:       oid.UUIDValue(),
			DisplayName: o.DisplayName(),
			Name:        o.Name().String(),
		}

		if org.Name != "" {
			if err := tx.db.Where("name = ?", org.Name).First(&scheme.Org{}).Error; err == nil {
				return sdkerrors.ErrAlreadyExists
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}

		return tx.db.Create(&org).Error
	})
	if err != nil {
		return sdktypes.InvalidOrgID, translateError(err)
	}

	return oid, nil
}

func (gdb *gormdb) UpdateOrg(ctx context.Context, o sdktypes.Org, fm *sdktypes.FieldMask) error {
	data, err := updatedFields(ctx, o, fm)
	if err != nil {
		return err
	}

	err = gdb.transaction(ctx, func(tx *tx) error {
		if name, ok := data["name"]; ok && name != "" {
			if err := tx.db.Where("name = ?", name).First(&scheme.Org{}).Error; err == nil {
				return sdkerrors.ErrAlreadyExists
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}

		return tx.db.
			Model(&scheme.Org{}).
			Where("org_id = ?", o.ID().UUIDValue()).
			Updates(data).
			Error
	})

	return translateError(err)
}

func (gdb *gormdb) ListOrgMembers(ctx context.Context, oid sdktypes.OrgID, includeUsers bool) ([]sdktypes.OrgMember, []sdktypes.User, error) {
	var mrs []scheme.OrgMember
	err := gdb.db.WithContext(ctx).
		Where("org_id = ?", oid.UUIDValue()).
		Order("created_at ASC").Find(&mrs).
		Error
	if err != nil {
		return nil, nil, translateError(err)
	}

	var urs []scheme.User

	if includeUsers {
		err = gdb.db.WithContext(ctx).
			Where("user_id in ?", kittehs.Transform(mrs, func(r scheme.OrgMember) uuid.UUID { return r.UserID })).
			Find(&urs).
			Error
		if err != nil {
			return nil, nil, translateError(err)
		}
	}

	ms, err := kittehs.TransformError(mrs, scheme.ParseOrgMember)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse org members: %w", err)
	}

	us, err := kittehs.TransformError(urs, scheme.ParseUser)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse users: %w", err)
	}

	return ms, us, nil
}

func (gdb *gormdb) AddOrgMember(ctx context.Context, m sdktypes.OrgMember) error {
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
	return translateError(
		gdb.db.WithContext(ctx).
			Where("org_id = ? AND user_id = ?", oid.UUIDValue(), uid.UUIDValue()).
			Delete(&scheme.OrgMember{}).
			Error,
	)
}

func (gdb *gormdb) GetOrgMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) (sdktypes.OrgMember, error) {
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

func (gdb *gormdb) GetOrgsForUser(ctx context.Context, uid sdktypes.UserID, includeOrgs bool) ([]sdktypes.OrgMember, []sdktypes.Org, error) {
	var rms []scheme.OrgMember
	err := gdb.db.WithContext(ctx).
		Where("user_id = ?", uid.UUIDValue()).
		Order("created_at ASC").
		Find(&rms).
		Error
	if err != nil {
		return nil, nil, translateError(err)
	}

	var ors []scheme.Org

	if includeOrgs {
		err = gdb.db.WithContext(ctx).
			Where("org_id in ?", kittehs.Transform(rms, func(r scheme.OrgMember) uuid.UUID { return r.OrgID })).
			Find(&ors).
			Error
		if err != nil {
			return nil, nil, translateError(err)
		}
	}

	ms, err := kittehs.TransformError(rms, scheme.ParseOrgMember)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse org members: %w", err)
	}

	os, err := kittehs.TransformError(ors, scheme.ParseOrg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse orgs: %w", err)
	}

	return ms, os, nil
}
