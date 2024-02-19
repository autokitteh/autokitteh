package dbgorm

import (
	"context"
	"slices"

	"go.autokitteh.dev/autokitteh/backend/internal/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (db *gormdb) SaveBuild(ctx context.Context, build sdktypes.Build, data []byte) error {
	// TODO: add Build time
	e := scheme.Build{
		BuildID:   sdktypes.GetBuildID(build).String(),
		ProjectID: sdktypes.GetBuildProjectID(build).String(),
		Data:      data,
		CreatedAt: sdktypes.GetBuildCreatedAt(build),
	}

	if err := db.db.WithContext(ctx).Create(&e).Error; err != nil {
		return translateError(err)
	}
	return nil
}

func (db *gormdb) GetBuild(ctx context.Context, buildID sdktypes.BuildID) (sdktypes.Build, error) {
	return get(db.db, ctx, scheme.ParseBuild, "build_id = ?", buildID.String())
}

func (db *gormdb) DeleteBuild(ctx context.Context, buildID sdktypes.BuildID) error {
	var b scheme.Build
	if err := db.db.WithContext(ctx).Where("build_id = ?", buildID.String()).Delete(&b).Error; err != nil {
		return translateError(err)
	}

	return nil
}

func (db *gormdb) ListBuilds(ctx context.Context, filter sdkservices.ListBuildsFilter) ([]sdktypes.Build, error) {
	q := db.db.WithContext(ctx)
	if filter.ProjectID != nil {
		q = q.Where("project_id = ?", filter.ProjectID.String())
	}

	q = q.Order("created_at desc")

	if filter.Limit != 0 {
		q = q.Limit(int(filter.Limit))
	}

	var bs []scheme.Build
	if err := q.Find(&bs).Error; err != nil {
		return nil, translateError(err)
	}

	slices.Reverse(bs)

	return kittehs.TransformError(bs, scheme.ParseBuild)
}

func (db *gormdb) GetBuildData(ctx context.Context, id sdktypes.BuildID) ([]byte, error) {
	var b scheme.Build
	if err := db.db.WithContext(ctx).Where("build_id = ?", id.String()).First(&b).Error; err != nil {
		return nil, translateError(err)
	}

	return b.Data, nil
}
