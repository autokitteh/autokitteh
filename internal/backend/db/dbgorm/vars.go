package dbgorm

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (gdb *gormdb) setVar(ctx context.Context, vr *scheme.Var) error {
	vr.VarID = vr.ScopeID // just ensure

	// TODO: Check that connection or trigger were not deleted.

	return gdb.transaction(ctx, func(tx *gormdb) error {
		db := tx.wdb

		var (
			tableName, idField string
			deletedAt          gorm.DeletedAt
		)

		if vr.IntegrationID == uuid.Nil {
			tableName, idField = "projects", "project_id"
		} else {
			tableName, idField = "connections", "connection_id"
		}

		query := fmt.Sprintf("SELECT deleted_at FROM %s where %s = ? LIMIT 1", tableName, idField)
		if err := db.Raw(query, vr.ScopeID).Scan(&deletedAt).Error; err != nil {
			return err
		}
		if deletedAt.Valid { // entity (either connection or project) was deleted
			return gorm.ErrForeignKeyViolated // emulate foreign keys check (#2)
		}

		return db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&vr).Error
	})
}

func varsCommonQuery(db *gorm.DB, scopeID uuid.UUID, names []string) *gorm.DB {
	db = db.Where("var_id = ?", scopeID)

	if len(names) > 0 {
		db = db.Where("name IN (?)", names)
	}
	return db
}

func (gdb *gormdb) deleteVars(ctx context.Context, scopeID uuid.UUID, names ...string) error {
	return varsCommonQuery(gdb.wdb.WithContext(ctx), scopeID, names).Delete(&scheme.Var{}).Error
}

func (gdb *gormdb) listVars(ctx context.Context, scopeID uuid.UUID, names ...string) ([]scheme.Var, error) {
	db := varsCommonQuery(gdb.rdb.WithContext(ctx), scopeID, names) // skip not user owned vars

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

func (gdb *gormdb) findConnectionIDsByVar(ctx context.Context, integrationID uuid.UUID, name string, v string) ([]uuid.UUID, error) {
	db := gdb.rdb.WithContext(ctx).Where("integration_id = ? AND name = ?", integrationID, name)
	if v != "" {
		db = db.Where("value = ? AND is_secret is false", v)
	}

	// Note(s):
	// - will skip not user owned vars
	// - not checking if scope is deleted, since scope deletion will cascade deletion of relevant vars
	var ids []uuid.UUID
	if err := db.Model(&scheme.Var{}).Distinct("var_id").Find(&ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

// FIXME: do we need to handle slice here? it seems to be unused anywhere and (meanwhile) compicates uniformal handling
func (db *gormdb) SetVars(ctx context.Context, vars []sdktypes.Var) error {
	i, err := kittehs.ValidateList(vars, func(_ int, v sdktypes.Var) error {
		return v.Strict()
	})
	if err != nil {
		return fmt.Errorf("var #%d: %w", i, err)
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
		var (
			sid = v.ScopeID()
			iid sdktypes.IntegrationID
		)

		if cid := sid.ToConnectionID(); cid.IsValid() {
			if iid = iids[cid]; !iid.IsValid() {
				return sdkerrors.NewInvalidArgumentError("integration id %v not found", iid)
			}
		} else if !sid.ToTriggerID().IsValid() && !sid.ToProjectID().IsValid() {
			return sdkerrors.NewInvalidArgumentError("unhandled scope %v", sid)
		}

		vr := scheme.Var{
			ScopeID:       sid.AsID().UUIDValue(), // scopeID is varID in DB
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

func (gdb *gormdb) CountVars(ctx context.Context, sid sdktypes.VarScopeID) (int, error) {
	q := varsCommonQuery(gdb.rdb.WithContext(ctx), sid.UUIDValue(), nil)

	var n int64
	if err := q.Model(&scheme.Var{}).Count(&n).Error; err != nil {
		return 0, translateError(err)
	}

	return int(n), nil
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

			return sdktypes.NewVar(n).SetValue(r.Value).SetSecret(r.IsSecret).WithScopeID(sid), nil
		},
	)
}

func (db *gormdb) FindConnectionIDsByVar(ctx context.Context, iid sdktypes.IntegrationID, n sdktypes.Symbol, v string) ([]sdktypes.ConnectionID, error) {
	if !iid.IsValid() {
		return nil, sdkerrors.NewInvalidArgumentError("integration id is invalid")
	}

	ids, err := db.findConnectionIDsByVar(ctx, iid.UUIDValue(), n.String(), v)
	if err != nil {
		return nil, translateError(err)
	}

	return kittehs.TransformError(ids, func(id uuid.UUID) (sdktypes.ConnectionID, error) {
		return sdktypes.NewIDFromUUID[sdktypes.ConnectionID](id), nil
	})
}
