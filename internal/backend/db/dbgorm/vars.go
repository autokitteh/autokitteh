package dbgorm

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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

	return db.transaction(ctx, func(tx *tx) error {
		dbvs := make([]scheme.Var, len(vars))

		for i, v := range vars {
			var iid, sid uuid.UUID

			if cid := v.ScopeID().ToConnectionID(); cid.IsValid() {
				sid = cid.UUIDValue()

				c, err := tx.GetConnection(ctx, cid)
				if err != nil {
					return err
				}

				iid = c.IntegrationID().UUIDValue()
			} else {
				sid = v.ScopeID().ToEnvID().UUIDValue()
			}

			dbvs[i] = scheme.Var{
				ScopeID:       sid,
				Name:          v.Name().String(),
				Value:         v.Value(),
				IsSecret:      v.IsSecret(),
				IntegrationID: iid,
			}
		}

		return translateError(tx.db.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&dbvs).Error)
	})
}

func (db *gormdb) GetVars(ctx context.Context, reveal bool, sid sdktypes.VarScopeID, ns []sdktypes.Symbol) ([]sdktypes.Var, error) {
	q := db.db.Where("scope_id = ?", sid.UUIDValue())

	if len(ns) > 0 {
		q = q.Where("name IN (?)", kittehs.TransformToStrings(ns))
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

func (db *gormdb) DeleteVars(ctx context.Context, sid sdktypes.VarScopeID, ns []sdktypes.Symbol) error {
	q := db.db.Where("scope_id = ?", sid.UUIDValue())

	if len(ns) > 0 {
		q = q.Where("name IN (?)", kittehs.TransformToStrings(ns))
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
	if err := q.Distinct("connection_id").Find(&vs).Error; err != nil {
		return nil, translateError(err)
	}

	return kittehs.TransformError(vs, func(r scheme.Var) (sdktypes.ConnectionID, error) {
		return sdktypes.NewIDFromUUID[sdktypes.ConnectionID](&r.ScopeID), nil
	})
}
