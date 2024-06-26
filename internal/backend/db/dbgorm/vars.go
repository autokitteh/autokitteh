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

func (gdb *gormdb) withUserVars(ctx context.Context) *gorm.DB {
	return gdb.withUserEntity(ctx, "var")
}

var varIDfunc = func() sdktypes.UUID { return sdktypes.NewVarID().UUIDValue() }

// FIXME:
// - env vars are user scopes for sure. Connection vars may have no user in ctx? Need to check more use-cases for connection vars

func (gdb *gormdb) setVar(ctx context.Context, vr *scheme.Var) error {
	conflict := clause.OnConflict{
		Columns:   []clause.Column{{Name: "scope_id"}, {Name: "name"}},      // uniqueness name + scope
		DoUpdates: clause.AssignmentColumns([]string{"value", "is_secret"}), // updates allowed for value and is_secret
	}

	// Set the ID for each var. Note that it be used only when creating the variable, on update with conflict it will be ignored
	vr.VarID = varIDfunc()

	// NOTE: createEntityWithOwnership won't respect clauses and it will enter to recirsive transaction
	// that's why we are handing user extraction, create entiry and ownerships differently here
	user, err := userFromContext(ctx)
	if err != nil {
		return err
	}

	return gdb.transaction(ctx, func(tx *tx) error {
		if err := tx.isUserEntity1(user, vr.ScopeID); err != nil {
			return err
		}

		// connection and env are soft-deleted. emulate foreign keys check
		// REVIEW: we may extract entityType from ownerships
		query := `
        SELECT CASE 
            WHEN EXISTS (SELECT 1 FROM connections WHERE connection_id = ? AND deleted_at IS NULL) THEN 1
            WHEN EXISTS (SELECT 1 FROM envs WHERE env_id = ? AND deleted_at IS NULL) THEN 1
            ELSE 0
        END`
		var count int64
		if err := tx.db.Raw(query, vr.ScopeID, vr.ScopeID).Scan(&count).Error; err != nil {
			return err
		}
		if count != 1 {
			return gorm.ErrForeignKeyViolated
		}

		ownerships := prepareOwnershipForEntities1(user, vr)
		if err := tx.db.Clauses(conflict).Create(vr).Error; err != nil {
			return err
		}
		return saveOwnershipForEntities(tx.db, user, ownerships...)
	})
}

func varsCommonQuery(db *gorm.DB, scopeID sdktypes.UUID, names []string) *gorm.DB {
	db = db.Where("scope_id = ?", scopeID)

	if len(names) > 0 {
		db = db.Where("name IN (?)", names)
	}
	return db
}

func (gdb *gormdb) deleteVars(ctx context.Context, scopeID sdktypes.UUID, names ...string) error {
	return gdb.transaction(ctx, func(tx *tx) error {
		if err := tx.isUserEntity(ctx, scopeID); err != nil {
			return err
		}
		return varsCommonQuery(tx.db, scopeID, names).Delete(&scheme.Var{}).Error
	})
}

func (gdb *gormdb) getVars(ctx context.Context, scopeID sdktypes.UUID, names ...string) ([]scheme.Var, error) {
	db := varsCommonQuery(gdb.withUserVars(ctx), scopeID, names)

	var vars []scheme.Var
	if err := db.Find(&vars).Error; err != nil {
		return nil, err
	}
	return vars, nil
}

func (gdb *gormdb) findConnectionIDsByVar(ctx context.Context, integrationID sdktypes.UUID, name string, v string) ([]scheme.Var, error) {
	db := gdb.withUserVars(ctx).Where("integration_id = ? AND name = ?", integrationID, name)
	if v != "" {
		db = db.Where("value = ? AND is_secret is false", v)
	}

	var vars []scheme.Var
	if err := db.Distinct("scope_id").Find(&vars).Error; err != nil {
		return nil, err
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
			ScopeID:       v.ScopeID().AsID().UUIDValue(),
			Name:          v.Name().String(),
			Value:         v.Value(),
			IsSecret:      v.IsSecret(),
			IntegrationID: iid.UUIDValue(),
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
