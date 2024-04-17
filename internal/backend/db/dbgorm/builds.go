package dbgorm

import (
	"context"
	"slices"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (db *gormdb) saveBuild(ctx context.Context, build *scheme.Build) error {
	return db.db.WithContext(ctx).Create(build).Error
}

func (db *gormdb) SaveBuild(ctx context.Context, build sdktypes.Build, data []byte) error {
	// TODO: add Build time
	b := scheme.Build{
		BuildID:   *build.ID().UUIDValue(),
		Data:      data,
		CreatedAt: build.CreatedAt(),
	}
	return translateError(db.saveBuild(ctx, &b))
}

func (db *gormdb) GetBuild(ctx context.Context, buildID sdktypes.BuildID) (sdktypes.Build, error) {
	return getOneWTransform(db.db, ctx, scheme.ParseBuild, "build_id = ?", buildID.UUIDValue())
}

func (db *gormdb) deleteBuild(ctx context.Context, buildID sdktypes.UUID) error {
	return delete(db.db, ctx, scheme.Build{}, "build_id = ?", buildID)
}

func (db *gormdb) DeleteBuild(ctx context.Context, buildID sdktypes.BuildID) error {
	return db.deleteBuild(ctx, *buildID.UUIDValue())
}

func (db *gormdb) listBuilds(ctx context.Context, filter sdkservices.ListBuildsFilter) ([]scheme.Build, error) {
	q := db.db.WithContext(ctx).Order("created_at desc")
	if filter.Limit != 0 {
		q = q.Limit(int(filter.Limit))
	}

	var bs []scheme.Build
	if err := q.Find(&bs).Error; err != nil {
		return nil, err
	}
	return bs, nil
}

func (db *gormdb) ListBuilds(ctx context.Context, filter sdkservices.ListBuildsFilter) ([]sdktypes.Build, error) {
	bs, err := db.listBuilds(ctx, filter)
	if err != nil {
		return nil, translateError(err)
	}

	slices.Reverse(bs)
	return kittehs.TransformError(bs, scheme.ParseBuild)
}

func (db *gormdb) GetBuildData(ctx context.Context, id sdktypes.BuildID) ([]byte, error) {
	var b scheme.Build
	if err := db.db.WithContext(ctx).Where("build_id = ?", id.UUIDValue()).First(&b).Error; err != nil {
		return nil, translateError(err)
	}

	return b.Data, nil
}
