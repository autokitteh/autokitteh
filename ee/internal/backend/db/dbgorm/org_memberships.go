package dbgorm

import (
	"context"
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"

	"go.autokitteh.dev/autokitteh/ee/internal/backend/db/dbgorm/scheme"
)

func orgMembershipID(oid sdktypes.OrgID, uid sdktypes.UserID) string {
	return fmt.Sprintf("%s/%s", oid.Value(), uid.Value())
}

func (db *gormdb) AddUserOrgMembership(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) error {
	r := scheme.OrgMember{
		MembershipID: orgMembershipID(oid, uid),
		OrgID:        oid.String(),
		UserID:       uid.String(),
	}

	if err := db.db.WithContext(ctx).Create(&r).Error; err != nil {
		return translateError(err)
	}

	return nil
}

func (db *gormdb) RemoveUserOrgMembership(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) error {
	err := db.db.WithContext(ctx).
		Where("membership_id = ?", orgMembershipID(oid, uid)).
		Delete(&scheme.OrgMember{}).Error
	if err != nil {
		return translateError(err)
	}

	return nil
}

func (db *gormdb) ListOrgMemberships(ctx context.Context, oid sdktypes.OrgID) ([]sdktypes.User, error) {
	var rs []scheme.User

	err := db.db.WithContext(ctx).
		Table("org_members").
		Joins("INNER JOIN users ON users.user_id == org_members.user_id").
		Where("org_members.org_id = ?", oid.String()).
		Select("users.user_id", "users.name").
		Order("users.user_id").
		Find(&rs).Error
	if err != nil {
		return nil, translateError(err)
	}

	return kittehs.TransformError(rs, func(r scheme.User) (sdktypes.User, error) {
		return sdktypes.StrictUserFromProto(&sdktypes.UserPB{
			UserId: r.UserID,
			Name:   r.Name,
		})
	})
}

func (db *gormdb) IsMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) (bool, error) {
	var ms []scheme.OrgMember

	mid := orgMembershipID(oid, uid)

	if err := translateError(
		db.db.WithContext(ctx).
			Where("membership_id = ?", mid).
			Limit(1).
			Find(&ms).
			Error,
	); err != nil {
		return false, err
	}

	return len(ms) == 1, nil
}

func (db *gormdb) ListUserOrgsMemberships(ctx context.Context, uid sdktypes.UserID) ([]sdktypes.Org, error) {
	type r struct {
		scheme.OrgMember
		scheme.Org
	}

	var rs []r

	err := db.db.WithContext(ctx).
		Table("org_members").
		Joins("INNER JOIN orgs ON orgs.org_id = org_members.org_id").
		Where("org_members.user_id = ?", uid.String()).
		Select("org_members.org_id", "orgs.name").
		Order("org_members.org_id").
		Find(&rs).Error
	if err != nil {
		return nil, translateError(err)
	}

	return kittehs.TransformError(rs, func(r r) (sdktypes.Org, error) {
		return sdktypes.StrictOrgFromProto(&sdktypes.OrgPB{
			OrgId: r.OrgMember.OrgID,
			Name:  r.Org.Name,
		})
	})
}
