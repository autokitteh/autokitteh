package dbgorm

import (
	"context"
	"fmt"

	"gorm.io/gorm/clause"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

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

	return db.transaction(ctx, func(tx *tx) error {
		cs, err := tx.GetConnections(ctx, cids)
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

		return translateError(tx.db.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&dbvs).Error)
	})
}

func (db *gormdb) GetVars(ctx context.Context, reveal bool, sid sdktypes.VarScopeID, names []sdktypes.Symbol) ([]sdktypes.Var, error) {
	q := db.db.Where("scope_id = ?", sid.UUIDValue())

	if len(names) > 0 {
		q = q.Where("name IN (?)", kittehs.TransformToStrings(names))
	}

	var dbvs []scheme.Var

	if err := q.Find(&dbvs).Error; err != nil {
		return nil, translateError(err)
	}

	return kittehs.TransformError(
		dbvs,
		func(r scheme.Var) (sdktypes.Var, error) {
			n, err := sdktypes.ParseSymbol(r.Name)
			if err != nil {
				return sdktypes.InvalidVar, err
			}

			if !reveal && r.IsSecret {
				r.Value = ""
			}

			return sdktypes.NewVar(n, r.Value, r.IsSecret).WithScopeID(sid), nil
		},
	)
}

func (db *gormdb) DeleteVars(ctx context.Context, sid sdktypes.VarScopeID, names []sdktypes.Symbol) error {
	q := db.db.Where("scope_id = ?", sid.UUIDValue())

	if len(names) > 0 {
		q = q.Where("name IN (?)", kittehs.TransformToStrings(names))
	}

	return translateError(q.Delete(&scheme.Var{}).Error)
}

func (db *gormdb) FindConnectionIDsByVar(ctx context.Context, iid sdktypes.IntegrationID, n sdktypes.Symbol, v string) ([]sdktypes.ConnectionID, error) {
	if !iid.IsValid() {
		return nil, sdkerrors.NewInvalidArgumentError("integration id is invalid")
	}

	q := db.db.Where("integration_id = ? AND name = ?", iid.UUIDValue(), n.String(), v)

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
