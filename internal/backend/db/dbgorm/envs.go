package dbgorm

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func envMembershipID(e sdktypes.Env) string {
	return fmt.Sprintf("%s/%s", e.ProjectID().Value(), e.Name().String())
}

func envVarMembershipID(ev sdktypes.EnvVar) string {
	return fmt.Sprintf("%s/%s", ev.EnvID().Value(), ev.Symbol().String())
}

func (db *gormdb) createEnv(ctx context.Context, env *scheme.Env) error {
	return db.db.WithContext(ctx).Create(env).Error
}

func (db *gormdb) CreateEnv(ctx context.Context, env sdktypes.Env) error {
	if !env.ID().IsValid() {
		db.z.DPanic("no env id supplied")
		return errors.New("env missing id")
	}

	e := scheme.Env{
		EnvID:        env.ID().String(),
		ProjectID:    scheme.PtrOrNil(env.ProjectID().String()), // TODO(ENG-136): need to verify parent id
		Name:         env.Name().String(),
		MembershipID: envMembershipID(env),
	}
	return translateError(db.createEnv(ctx, &e))
}

func (db *gormdb) deleteEnvs(ctx context.Context, ids []string) error {
	// NOTE: should be transactional
	gormDB := db.db.WithContext(ctx)

	// enforce foreign keys constrains while soft-deleting
	var count int64
	gormDB.Model(&scheme.Deployment{}).Where("deleted_at is NULL and env_id IN ?", ids).Count(&count)
	if count > 0 {
		return fmt.Errorf("FOREIGN KEY: %w", gorm.ErrForeignKeyViolated)
	}

	return gormDB.Where("env_id IN ?", ids).Delete(&scheme.Env{}).Error
}

func (db *gormdb) deleteEnv(ctx context.Context, envID string) error {
	return db.transaction(ctx, func(tx *tx) error {
		return tx.deleteEnvs(ctx, []string{envID})
	})
}

func (db *gormdb) GetEnvByID(ctx context.Context, eid sdktypes.EnvID) (sdktypes.Env, error) {
	return getOneWTransform(db.db, ctx, scheme.ParseEnv, "env_id = ?", eid.String())
}

func (db *gormdb) GetEnvByName(ctx context.Context, pid sdktypes.ProjectID, h sdktypes.Symbol) (sdktypes.Env, error) {
	return getOneWTransform(db.db, ctx, scheme.ParseEnv, "project_id = ? AND name = ?", pid.String(), h.String())
}

func (db *gormdb) ListProjectEnvs(ctx context.Context, pid sdktypes.ProjectID) ([]sdktypes.Env, error) {
	var rs []scheme.Env

	q := db.db.WithContext(ctx).Order("env_id")

	if pid.IsValid() {
		q = q.Where("project_id = ?", pid.String())
	}

	err := q.Find(&rs).Error
	if err != nil {
		return nil, translateError(err)
	}

	return kittehs.TransformError(rs, func(r scheme.Env) (sdktypes.Env, error) {
		return sdktypes.StrictEnvFromProto(&sdktypes.EnvPB{
			EnvId:     r.EnvID,
			ProjectId: *r.ProjectID,
			Name:      r.Name,
		})
	})
}

func (db *gormdb) SetEnvVar(ctx context.Context, ev sdktypes.EnvVar) error {
	r := scheme.EnvVar{
		EnvID:        ev.EnvID().String(), // need to verify envID ? where is envvar id ?
		Name:         ev.Symbol().String(),
		IsSecret:     ev.IsSecret(),
		MembershipID: envVarMembershipID(ev),
	}

	if r.IsSecret {
		r.SecretValue = ev.Value()
	} else {
		r.Value = ev.Value()
	}

	if err := db.db.
		WithContext(ctx).
		Clauses(clause.OnConflict{UpdateAll: true, Columns: []clause.Column{{Name: "membership_id"}}}). // upsert.
		Create(&r).Error; err != nil {
		return translateError(err)
	}

	return nil
}

func (db *gormdb) GetEnvVars(ctx context.Context, eid sdktypes.EnvID) ([]sdktypes.EnvVar, error) {
	var rs []scheme.EnvVar
	err := db.db.WithContext(ctx).
		Where("env_id = ?", eid.String()).
		Select("env_id", "name", "value", "is_secret"). // exclude secret_value.
		Order("name").
		Find(&rs).Error
	if err != nil {
		return nil, translateError(err)
	}

	return kittehs.TransformError(rs, scheme.ParseEnvVar)
}

func (db *gormdb) RevealEnvVar(ctx context.Context, eid sdktypes.EnvID, vn sdktypes.Symbol) (string, error) {
	var r scheme.EnvVar
	if err := db.db.WithContext(ctx).Where("env_id = ? and name = ?", eid.String(), vn.String()).First(&r).Error; err != nil {
		return "", translateError(err)
	}

	if r.IsSecret {
		return r.SecretValue, nil
	}

	return r.Value, nil
}
