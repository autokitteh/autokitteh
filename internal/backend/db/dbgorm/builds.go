package dbgorm

import (
	"context"
	"slices"

	"github.com/google/uuid"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (gdb *gormdb) saveBuild(ctx context.Context, build *scheme.Build) error {
	return gdb.wdb.WithContext(ctx).Create(build).Error
}

func (gdb *gormdb) deleteBuild(ctx context.Context, buildID uuid.UUID) error {
	return gdb.wdb.WithContext(ctx).Delete(&scheme.Build{BuildID: buildID}).Error
}

func (gdb *gormdb) getBuild(ctx context.Context, buildID uuid.UUID) (*scheme.Build, error) {
	return getOne[scheme.Build](gdb.rdb.WithContext(ctx), "build_id = ?", buildID)
}

func (gdb *gormdb) listBuilds(ctx context.Context, filter sdkservices.ListBuildsFilter) ([]scheme.Build, error) {
	q := gdb.rdb.WithContext(ctx).Order("created_at desc")

	q = withProjectID(q, "", filter.ProjectID)

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

	b := scheme.Build{
		Base:      based(ctx),
		ProjectID: build.ProjectID().UUIDValue(),
		BuildID:   build.ID().UUIDValue(),
		Data:      data,
	}

	return translateError(db.saveBuild(ctx, &b))
}

func (db *gormdb) DeleteBuild(ctx context.Context, buildID sdktypes.BuildID) error {
	return translateError(db.deleteBuild(ctx, buildID.UUIDValue()))
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
	builds, err := db.listBuilds(ctx, filter)
	if builds == nil || err != nil {
		return nil, translateError(err)
	}

	slices.Reverse(builds)
	return kittehs.TransformError(builds, scheme.ParseBuild)
}
