package dbgorm

import (
	"context"
	"slices"

	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (gdb *gormdb) withUserBuilds(ctx context.Context) *gorm.DB {
	return gdb.withUserEntity(ctx, "build")
}

func (gdb *gormdb) saveBuild(ctx context.Context, build *scheme.Build) error {
	createFunc := func(tx *gorm.DB, uid string) error { return tx.Create(build).Error }
	return gdb.createEntityWithOwnership(ctx, createFunc, build, build.ProjectID)
}

func (gdb *gormdb) deleteBuild(ctx context.Context, buildID sdktypes.UUID) error {
	return gdb.transaction(ctx, func(tx *tx) error {
		if err := tx.isCtxUserEntity(ctx, buildID); err != nil {
			return err
		}
		return tx.db.Delete(&scheme.Build{BuildID: buildID}).Error
	})
}

func (gdb *gormdb) getBuild(ctx context.Context, buildID sdktypes.UUID) (*scheme.Build, error) {
	return getOne[scheme.Build](gdb.withUserBuilds(ctx), "build_id = ?", buildID)
}

func (gdb *gormdb) listBuilds(ctx context.Context, filter sdkservices.ListBuildsFilter) ([]scheme.Build, error) {
	q := gdb.withUserBuilds(ctx).Order("created_at desc")

	if filter.Limit != 0 {
		q = q.Limit(int(filter.Limit))
	}

	var bs []scheme.Build
	if err := q.Find(&bs).Error; err != nil {
		return nil, err
	}
	return bs, nil
}

func (db *gormdb) SaveBuild(ctx context.Context, build sdktypes.Build, data []byte) error {
	if err := build.Strict(); err != nil {
		return err
	}

	// TODO: add Build time
	b := scheme.Build{
		BuildID:   build.ID().UUIDValue(),
		ProjectID: scheme.UUIDOrNil(build.ProjectID().UUIDValue()),
		Data:      data,
		CreatedAt: build.CreatedAt(),
	}
	return translateError(db.saveBuild(ctx, &b))
}

func (db *gormdb) DeleteBuild(ctx context.Context, buildID sdktypes.BuildID) error {
	return translateError(db.deleteBuild(ctx, buildID.UUIDValue()))
}

func (db *gormdb) GetBuild(ctx context.Context, buildID sdktypes.BuildID) (sdktypes.Build, error) {
	b, err := db.getBuild(ctx, buildID.UUIDValue())
	return schemaToSDK(b, err, scheme.ParseBuild)
}

func (db *gormdb) GetBuildData(ctx context.Context, buildID sdktypes.BuildID) ([]byte, error) {
	b, err := db.getBuild(ctx, buildID.UUIDValue())
	if err != nil {
		return nil, translateError(err)
	}
	return b.Data, nil
}

func (db *gormdb) ListBuilds(ctx context.Context, filter sdkservices.ListBuildsFilter) ([]sdktypes.Build, error) {
	bs, err := db.listBuilds(ctx, filter)
	builds, err := schemasToSDK(bs, err, scheme.ParseBuild)
	slices.Reverse(builds)
	return builds, err
}
