package dbgorm

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm/clause"

	"go.autokitteh.dev/autokitteh/backend/internal/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func envMembershipID(e sdktypes.Env) string {
	return fmt.Sprintf("%s/%s", sdktypes.GetEnvProjectID(e).Value(), sdktypes.GetEnvName(e).String())
}

func envVarMembershipID(ev sdktypes.EnvVar) string {
	return fmt.Sprintf("%s/%s", sdktypes.GetEnvVarEnvID(ev).Value(), sdktypes.GetEnvVarName(ev).String())
}

func (db *gormdb) CreateEnv(ctx context.Context, env sdktypes.Env) error {
	if !sdktypes.EnvHasID(env) {
		db.z.DPanic("no env id supplied")
		return errors.New("env missing id")
	}

	r := scheme.Env{
		EnvID:        sdktypes.GetEnvID(env).String(),
		ProjectID:    sdktypes.GetEnvProjectID(env).String(), // TODO(ENG-136): need to verify parent id
		Name:         sdktypes.GetEnvName(env).String(),
		MembershipID: envMembershipID(env),
	}

	if err := db.db.WithContext(ctx).Create(&r).Error; err != nil {
		return translateError(err)
	}

	return nil
}

func (db *gormdb) GetEnvByID(ctx context.Context, eid sdktypes.EnvID) (sdktypes.Env, error) {
	return get(db.db, ctx, scheme.ParseEnv, "env_id = ?", eid.String())
}

func (db *gormdb) GetEnvByName(ctx context.Context, pid sdktypes.ProjectID, h sdktypes.Name) (sdktypes.Env, error) {
	return get(db.db, ctx, scheme.ParseEnv, "project_id = ? AND name = ?", pid.String(), h.String())
}

func (db *gormdb) ListProjectEnvs(ctx context.Context, pid sdktypes.ProjectID) ([]sdktypes.Env, error) {
	var rs []scheme.Env

	q := db.db.WithContext(ctx).Order("env_id")

	if pid != nil {
		q = q.Where("project_id = ?", pid.String())
	}

	err := q.Find(&rs).Error
	if err != nil {
		return nil, translateError(err)
	}

	return kittehs.TransformError(rs, func(r scheme.Env) (sdktypes.Env, error) {
		return sdktypes.StrictEnvFromProto(&sdktypes.EnvPB{
			EnvId:     r.EnvID,
			ProjectId: r.ProjectID,
			Name:      r.Name,
		})
	})
}

func (db *gormdb) SetEnvVar(ctx context.Context, ev sdktypes.EnvVar) error {
	r := scheme.EnvVar{
		EnvID:        sdktypes.GetEnvVarEnvID(ev).String(), // need to verify envID ? where is envvar id ?
		Name:         sdktypes.GetEnvVarName(ev).String(),
		IsSecret:     sdktypes.IsEnvVarSecret(ev),
		MembershipID: envVarMembershipID(ev),
	}

	if r.IsSecret {
		r.SecretValue = sdktypes.GetEnvVarValue(ev)
	} else {
		r.Value = sdktypes.GetEnvVarValue(ev)
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
