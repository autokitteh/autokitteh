package dbgorm

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func envMembershipID(e sdktypes.Env) string {
	return fmt.Sprintf("%s/%s", e.ProjectID().UUIDValue(), e.Name().String())
}

func (db *gormdb) createEnv(ctx context.Context, env *scheme.Env) error {
	return db.db.WithContext(ctx).Create(env).Error
}

func (gdb *gormdb) createEnvWithOwnership(ctx context.Context, env *scheme.Env) error {
	createFunc := func(p *scheme.Env) error { return gdb.createEnv(ctx, env) }
	return createEntityWithOwnership(ctx, gdb, env, createFunc)
}

func (db *gormdb) CreateEnv(ctx context.Context, env sdktypes.Env) error {
	if err := env.Strict(); err != nil {
		return err
	}

	e := scheme.Env{
		EnvID:        env.ID().UUIDValue(),
		ProjectID:    env.ProjectID().UUIDValue(),
		Name:         env.Name().String(),
		MembershipID: envMembershipID(env),
	}
	return translateError(db.createEnvWithOwnership(ctx, &e))
}

func (db *gormdb) deleteEnvs(ctx context.Context, ids []sdktypes.UUID) error {
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

func (db *gormdb) deleteEnv(ctx context.Context, envID sdktypes.UUID) error {
	return db.transaction(ctx, func(tx *tx) error {
		return tx.deleteEnvs(ctx, []sdktypes.UUID{envID})
	})
}

func (db *gormdb) GetEnvByID(ctx context.Context, eid sdktypes.EnvID) (sdktypes.Env, error) {
	return getOneWTransform(db.db, ctx, scheme.ParseEnv, "env_id = ?", eid.UUIDValue())
}

func (db *gormdb) GetEnvByName(ctx context.Context, pid sdktypes.ProjectID, h sdktypes.Symbol) (sdktypes.Env, error) {
	return getOneWTransform(db.db, ctx, scheme.ParseEnv, "project_id = ? AND name = ?", pid.UUIDValue(), h.String())
}

func (db *gormdb) ListProjectEnvs(ctx context.Context, pid sdktypes.ProjectID) ([]sdktypes.Env, error) {
	var rs []scheme.Env

	q := db.db.WithContext(ctx).Order("env_id")

	if pid.IsValid() {
		q = q.Where("project_id = ?", pid.UUIDValue())
	}

	err := q.Find(&rs).Error
	if err != nil {
		return nil, translateError(err)
	}

	return kittehs.TransformError(rs, func(r scheme.Env) (sdktypes.Env, error) {
		return sdktypes.StrictEnvFromProto(&sdktypes.EnvPB{
			EnvId:     sdktypes.NewIDFromUUID[sdktypes.EnvID](&r.EnvID).String(),
			ProjectId: sdktypes.NewIDFromUUID[sdktypes.ProjectID](&r.ProjectID).String(),
			Name:      r.Name,
		})
	})
}
