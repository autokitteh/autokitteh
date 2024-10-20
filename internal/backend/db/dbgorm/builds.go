package dbgorm

import (
	"context"
	"slices"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (gdb *gormdb) getBuild(ctx context.Context, buildID sdktypes.UUID) (*scheme.Build, error) {
	return getOne[scheme.Build](gdb.db.WithContext(ctx), "build_id = ?", buildID)
}

func (db *gormdb) SaveBuild(ctx context.Context, build sdktypes.Build, data []byte) error {
	if err := build.Strict(); err != nil {
		return err
	}

	b := scheme.Build{
		BuildID:   build.ID().UUIDValue(),
		ProjectID: scheme.UUIDOrNil(build.ProjectID().UUIDValue()),
		Data:      data,
		CreatedAt: build.CreatedAt(),
	}

	return translateError(db.db.Create(b).Error)
}

func (db *gormdb) DeleteBuild(ctx context.Context, buildID sdktypes.BuildID) error {
	return translateError(db.db.Delete(&scheme.Build{BuildID: buildID.UUIDValue()}).Error)
}

func (db *gormdb) GetBuild(ctx context.Context, buildID sdktypes.BuildID) (sdktypes.Build, error) {
	b, err := db.getBuild(ctx, buildID.UUIDValue())
	if b == nil || err != nil {
		return sdktypes.InvalidBuild, translateError(err)
	}
	return scheme.ParseBuild(*b) // TODO: get and list returns different errors due to transform
}

func (db *gormdb) GetBuildData(ctx context.Context, buildID sdktypes.BuildID) ([]byte, error) {
	b, err := db.getBuild(ctx, buildID.UUIDValue())
	if b == nil || err != nil {
		return nil, translateError(err)
	}
	return b.Data, nil
}

func (db *gormdb) ListBuilds(ctx context.Context, filter sdkservices.ListBuildsFilter) ([]sdktypes.Build, error) {
	q := db.db.Order("created_at desc")

	if filter.Limit != 0 {
		q = q.Limit(int(filter.Limit))
	}

	var bs []scheme.Build
	if err := q.Find(&bs).Error; err != nil {
		return nil, err
	}

	slices.Reverse(bs)

	return kittehs.TransformError(bs, scheme.ParseBuild)
}
