package dbgorm

import (
	"context"
	"errors"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"

	"go.autokitteh.dev/autokitteh/ee/internal/backend/db/dbgorm/scheme"
)

func (db *gormdb) CreateOrg(ctx context.Context, org sdktypes.Org) error {
	if !org.ID().IsValid() {
		db.z.DPanic("no org id supplied")
		return errors.New("org id missing")
	}

	r := scheme.Org{
		OrgID: org.ID().String(),
		Name:  org.Name().String(),
	}

	if err := db.db.WithContext(ctx).Create(&r).Error; err != nil {
		return translateError(err)
	}

	return nil
}

func (db *gormdb) GetOrgByID(ctx context.Context, oid sdktypes.OrgID) (sdktypes.Org, error) {
	var r scheme.Org
	if err := db.db.WithContext(ctx).Where("org_id = ?", oid.String()).First(&r).Error; err != nil {
		return sdktypes.InvalidOrg, translateError(err)
	}

	return scheme.ParseOrg(r)
}

func (db *gormdb) GetOrgByName(ctx context.Context, name sdktypes.Symbol) (sdktypes.Org, error) {
	return get(db.db, ctx, scheme.ParseOrg, "name = ?", name.String())
}
