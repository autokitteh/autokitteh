package dbgorm

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (gdb *gormdb) withUserEnvs(ctx context.Context) *gorm.DB {
	return gdb.withUserEntity(ctx, "env")
}

func (gdb *gormdb) createEnv(ctx context.Context, env *scheme.Env) error {
	createFunc := func(tx *gorm.DB, uid string) error { return tx.Create(env).Error }
	return gormErrNotFoundToForeignKey(gdb.createEntityWithOwnership(ctx, createFunc, env, &env.ProjectID))
}

func (gdb *gormdb) deleteEnv(ctx context.Context, envID sdktypes.UUID) error {
	return gdb.transaction(ctx, func(tx *tx) error {
		return tx.deleteEnvs(ctx, []sdktypes.UUID{envID})
	})
}

func (gdb *gormdb) deleteEnvs(ctx context.Context, ids []sdktypes.UUID) error {
	// NOTE: should be transactional
	db := gdb.db.WithContext(ctx)

	// enforce foreign keys constrains while soft-deleting - should be no active deployments
	var count int64
	db.Model(&scheme.Deployment{}).Where("deleted_at is NULL and env_id IN ?", ids).Count(&count)
	if count > 0 {
		return fmt.Errorf("FOREIGN KEY: %w", gorm.ErrForeignKeyViolated)
	}

	// REVIEW: should we allow deleting only user envs and skipping the others.
	// The check below will fail of any of provided envs cannot be deleted
	if err := gdb.isCtxUserEntity(ctx, ids...); err != nil {
		return err
	}

	// delete envs and associated vars
	if err := db.Where("env_id IN ?", ids).Delete(&scheme.Env{}).Error; err != nil {
		return err
	}

	return db.Where("var_id IN ?", ids).Delete(&scheme.Var{}).Error
}

func (gdb *gormdb) getEnvByID(ctx context.Context, envID sdktypes.UUID) (*scheme.Env, error) {
	return getOne[scheme.Env](gdb.withUserEnvs(ctx), "env_id = ?", envID)
}

func (gdb *gormdb) getEnvByName(ctx context.Context, projectID sdktypes.UUID, envName string) (*scheme.Env, error) {
	return getOne[scheme.Env](gdb.withUserEnvs(ctx), "project_id = ? AND name = ?", projectID, envName)
}

func (gdb *gormdb) listEnvs(ctx context.Context, projectID sdktypes.UUID) ([]scheme.Env, error) {
	// REVIEW: project could belong to someone else. Just fetch envs belong to user

	var envs []scheme.Env
	q := gdb.withUserEnvs(ctx).Where("project_id = ?", projectID).Order("env_id")
	if err := q.Find(&envs).Error; err != nil {
		return nil, err
	}
	return envs, nil
}

func (db *gormdb) CreateEnv(ctx context.Context, env sdktypes.Env) error {
	if err := env.Strict(); err != nil {
		return err
	}

	envName := env.Name().String()
	projectID := env.ProjectID().UUIDValue()
	e := scheme.Env{
		EnvID:        env.ID().UUIDValue(),
		ProjectID:    projectID,
		Name:         envName,
		MembershipID: fmt.Sprintf("%s/%s", projectID, envName), // ensure no duplicate env name in the same project
	}
	return translateError(db.createEnv(ctx, &e))
}

func (db *gormdb) GetEnvByID(ctx context.Context, envID sdktypes.EnvID) (sdktypes.Env, error) {
	e, err := db.getEnvByID(ctx, envID.UUIDValue())
	return schemaToSDK(e, err, scheme.ParseEnv)
}

func (db *gormdb) GetEnvByName(ctx context.Context, projectID sdktypes.ProjectID, envName sdktypes.Symbol) (sdktypes.Env, error) {
	e, err := db.getEnvByName(ctx, projectID.UUIDValue(), envName.String())
	return schemaToSDK(e, err, scheme.ParseEnv)
}

func (db *gormdb) ListProjectEnvs(ctx context.Context, pid sdktypes.ProjectID) ([]sdktypes.Env, error) {
	envs, err := db.listEnvs(ctx, pid.UUIDValue())
	return schemasToSDK(envs, err, scheme.ParseEnv)
}
