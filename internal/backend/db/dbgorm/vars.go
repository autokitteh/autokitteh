package dbgorm

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (db *gormdb) setVars(ctx context.Context, vars []scheme.Var) error {
	// NOTE: should be transactional
	return db.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(&vars).Error
}

func (db *gormdb) SetVars(ctx context.Context, vars []sdktypes.Var) error {
	if i, err := kittehs.ValidateList(vars, func(_ int, v sdktypes.Var) error {
		return v.Strict()
	}); err != nil {
		return fmt.Errorf("#%d: %w", i, err)
	}

	cids := kittehs.Transform(vars, func(v sdktypes.Var) sdktypes.ConnectionID {
		return v.ScopeID().ToConnectionID()
	})

	cids = kittehs.Filter(cids, func(cid sdktypes.ConnectionID) bool {
		return cid.IsValid()
	})

	// NOTE: should be transactional  ---v
	cs, err := db.GetConnections(ctx, cids)
	if err != nil {
		return err
	}

	iids := kittehs.ListToMap(cs, func(c sdktypes.Connection) (sdktypes.ConnectionID, sdktypes.IntegrationID) {
		return c.ID(), c.IntegrationID()
	})

	dbvs := make([]scheme.Var, len(vars))

	for i, v := range vars {
		var iid sdktypes.IntegrationID

		if cid := v.ScopeID().ToConnectionID(); cid.IsValid() {
			if iid = iids[cid]; !iid.IsValid() {
				return sdkerrors.NewInvalidArgumentError("integration id %v not found", iid)
			}
		}

		dbvs[i] = scheme.Var{
			ScopeID:       v.ScopeID().AsID().UUIDValue(),
			Name:          v.Name().String(),
			Value:         v.Value(),
			IsSecret:      v.IsSecret(),
			IntegrationID: iid.UUIDValue(),
		}
	}

	return translateError(db.setVars(ctx, dbvs))
}

func (db *gormdb) varsCommonQuery(ctx context.Context, scopeID sdktypes.UUID, names []string) *gorm.DB {
	gormDB := db.db.WithContext(ctx)
	query := gormDB.Where("scope_id = ?", scopeID)

	if len(names) > 0 {
		query = query.Where("name IN (?)", names)
	}
	return query
}

func (db *gormdb) getVars(ctx context.Context, scopeID sdktypes.UUID, names ...string) ([]scheme.Var, error) {
	query := db.varsCommonQuery(ctx, scopeID, names)

	var vars []scheme.Var

	if err := query.Find(&vars).Error; err != nil {
		return nil, err
	}
	return vars, nil
}

func (db *gormdb) GetVars(ctx context.Context, sid sdktypes.VarScopeID, names []sdktypes.Symbol) ([]sdktypes.Var, error) {
	dbvs, err := db.getVars(ctx, sid.UUIDValue(), kittehs.TransformToStrings(names)...)
	if err != nil {
		return nil, translateError(err)
	}

	return kittehs.TransformError(
		dbvs,
		func(r scheme.Var) (sdktypes.Var, error) {
			n, err := sdktypes.ParseSymbol(r.Name)
			if err != nil {
				return sdktypes.InvalidVar, err
			}

			return sdktypes.NewVar(n, r.Value, r.IsSecret).WithScopeID(sid), nil
		},
	)
}

func (db *gormdb) deleteVars(ctx context.Context, scopeID sdktypes.UUID, names ...string) error {
	query := db.varsCommonQuery(ctx, scopeID, names)
	return query.Delete(&scheme.Var{}).Error
}

func (db *gormdb) DeleteVars(ctx context.Context, sid sdktypes.VarScopeID, names []sdktypes.Symbol) error {
	return translateError(db.deleteVars(ctx, sid.UUIDValue(), kittehs.TransformToStrings(names)...))
}

func (db *gormdb) FindConnectionIDsByVar(ctx context.Context, iid sdktypes.IntegrationID, n sdktypes.Symbol, v string) ([]sdktypes.ConnectionID, error) {
	if !iid.IsValid() {
		return nil, sdkerrors.NewInvalidArgumentError("integration id is invalid")
	}

	q := db.db.Where("integration_id = ? AND name = ?", iid.UUIDValue(), n.String())

	if v != "" {
		q = q.Where("value = ? AND is_secret is false", v)
	}

	var vs []scheme.Var
	if err := q.Distinct("scope_id").Find(&vs).Error; err != nil {
		return nil, translateError(err)
	}

	return kittehs.TransformError(vs, func(r scheme.Var) (sdktypes.ConnectionID, error) {
		return sdktypes.NewIDFromUUID[sdktypes.ConnectionID](&r.ScopeID), nil
	})
}
