package dbgorm

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (gdb *gormdb) withUserVars(ctx context.Context) *gorm.DB {
	return gdb.withUserEntity(ctx, "var")
}

func (gdb *gormdb) setVar(ctx context.Context, vr *scheme.Var) error {
	vr.VarID = vr.ScopeID // just ensure

	uid, err := userIDFromContext(ctx)
	if err != nil {
		return err
	}
	return gdb.transaction(ctx, func(tx *tx) error {
		db := tx.db.WithContext(ctx)

		// if no records were found then fail with foreign keys validation (#1), since there should be one
		oo, err := tx.owner.EnsureUserAccessToEntitiesWithOwnership(ctx, db, uid, vr.ScopeID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return gorm.ErrForeignKeyViolated
			}
			return err
		}
		if len(oo) != 1 {
			return gorm.ErrForeignKeyViolated
		}

		var tableName, idField string
		var deletedAt gorm.DeletedAt
		switch oo[0].EntityType {
		case "Connection":
			tableName, idField = "connections", "connection_id"
		case "Env":
			tableName, idField = "envs", "env_id"
		default:
			return gorm.ErrCheckConstraintViolated // should be either Env or Connection
		}
		query := fmt.Sprintf("SELECT deleted_at FROM %s where %s = ? LIMIT 1", tableName, idField)
		if err := db.Raw(query, vr.ScopeID).Scan(&deletedAt).Error; err != nil {
			return err
		}
		if deletedAt.Valid { // entity (either connection or env) was deleted
			return gorm.ErrForeignKeyViolated // emulate foreign keys check (#2)
		}

		return db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&vr).Error
	})
}

func varsCommonQuery(db *gorm.DB, scopeID sdktypes.UUID, names []string) *gorm.DB {
	db = db.Where("var_id = ?", scopeID)

	if len(names) > 0 {
		db = db.Where("name IN (?)", names)
	}
	return db
}

func (gdb *gormdb) deleteVars(ctx context.Context, scopeID sdktypes.UUID, names ...string) error {
	return gdb.transaction(ctx, func(tx *tx) error {
		if err := tx.isCtxUserEntity(ctx, scopeID); err != nil {
			return err
		}
		return varsCommonQuery(tx.db, scopeID, names).Delete(&scheme.Var{}).Error
	})
}

func (gdb *gormdb) listVars(ctx context.Context, scopeID sdktypes.UUID, names ...string) ([]scheme.Var, error) {
	db := varsCommonQuery(gdb.withUserVars(ctx), scopeID, names) // skip not user owned vars

	// Note: we are not checking if scope (env/connection) is deleted, since scope deletion will cascade deletion of relevant vars
	// e.g. vars present only for valid and active scope
	var vars []scheme.Var
	if err := db.Find(&vars).Error; err != nil {
		return nil, err
	}
	for i := range vars {
		vars[i].ScopeID = vars[i].VarID
	}
	return vars, nil
}

func (gdb *gormdb) findConnectionIDsByVar(ctx context.Context, integrationID sdktypes.UUID, name string, v string) ([]scheme.Var, error) {
	db := gdb.withUserVars(ctx).Where("integration_id = ? AND name = ?", integrationID, name)
	if v != "" {
		db = db.Where("value = ? AND is_secret is false", v)
	}

	// Note(s):
	// - will skip not user owned vars
	// - not checking if scope is deleted, since scope deletion will cascade deletion of relevant vars
	var vars []scheme.Var
	if err := db.Distinct("var_id").Find(&vars).Error; err != nil {
		return nil, err
	}
	for i := range vars {
		vars[i].ScopeID = vars[i].VarID
	}
	return vars, nil
}

// FIXME: do we need to handle slice here? it seems to be unused anywhere and (meanwhile) compicates uniformal handling
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

	for _, v := range vars {
		var iid sdktypes.IntegrationID

		if cid := v.ScopeID().ToConnectionID(); cid.IsValid() {
			if iid = iids[cid]; !iid.IsValid() {
				return sdkerrors.NewInvalidArgumentError("integration id %v not found", iid)
			}
		}

		vr := scheme.Var{
			ScopeID:       v.ScopeID().AsID().UUIDValue(), // scopeID is varID in DB
			Name:          v.Name().String(),
			Value:         v.Value(),
			IsSecret:      v.IsSecret(),
			IsOptional:    v.IsOptional(),
			IntegrationID: iid.UUIDValue(),
			Description:   v.Description(),
		}
		if err := db.setVar(ctx, &vr); err != nil {
			return translateError(err)
		}
	}
	return nil
}

func (db *gormdb) DeleteVars(ctx context.Context, sid sdktypes.VarScopeID, names []sdktypes.Symbol) error {
	return translateError(db.deleteVars(ctx, sid.UUIDValue(), kittehs.TransformToStrings(names)...))
}

func (db *gormdb) GetVars(ctx context.Context, sid sdktypes.VarScopeID, names []sdktypes.Symbol) ([]sdktypes.Var, error) {
	dbvs, err := db.listVars(ctx, sid.UUIDValue(), kittehs.TransformToStrings(names)...)
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

			return sdktypes.NewVar(n).SetValue(r.Value).SetSecret(r.IsSecret).SetOptional(r.IsOptional).WithScopeID(sid).SetDescription(r.Description), nil
		},
	)
}

func (db *gormdb) FindConnectionIDsByVar(ctx context.Context, iid sdktypes.IntegrationID, n sdktypes.Symbol, v string) ([]sdktypes.ConnectionID, error) {
	if !iid.IsValid() {
		return nil, sdkerrors.NewInvalidArgumentError("integration id is invalid")
	}

	vs, err := db.findConnectionIDsByVar(ctx, iid.UUIDValue(), n.String(), v)
	if err != nil {
		return nil, translateError(err)
	}

	return kittehs.TransformError(vs, func(r scheme.Var) (sdktypes.ConnectionID, error) {
		return sdktypes.NewIDFromUUID[sdktypes.ConnectionID](&r.ScopeID), nil
	})
}
