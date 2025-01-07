package builds

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authz"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/telemetry"
	"go.autokitteh.dev/autokitteh/sdk/sdkbuildfile"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Builds struct {
	fx.In

	Z  *zap.Logger
	DB db.DB
}

func New(b Builds, telemetry *telemetry.Telemetry) sdkservices.Builds {
	initMetrics(telemetry)
	return &b
}

func (b *Builds) Save(ctx context.Context, build sdktypes.Build, data []byte) (sdktypes.BuildID, error) {
	if err := authz.CheckContext(
		ctx,
		sdktypes.InvalidBuildID,
		"create:save",
		authz.WithData("build", build),
		authz.WithAssociationWithID("project", build.ProjectID()),
	); err != nil {
		return sdktypes.InvalidBuildID, err
	}

	// make sure this at least tries to pretend to be a build file.
	if _, err := sdkbuildfile.ReadVersion(bytes.NewReader(data)); err != nil {
		return sdktypes.InvalidBuildID, fmt.Errorf("read version: %w", err)
	}

	build = build.WithNewID().WithCreatedAt(time.Now())

	if err := b.DB.SaveBuild(ctx, build, data); err != nil {
		return sdktypes.InvalidBuildID, err
	}

	buildsCreatedCounter.Add(ctx, 1)
	return build.ID(), nil
}

func (b *Builds) Get(ctx context.Context, id sdktypes.BuildID) (sdktypes.Build, error) {
	if err := authz.CheckContext(ctx, id, "read:get", authz.WithConvertForbiddenToNotFound); err != nil {
		return sdktypes.InvalidBuild, err
	}

	return b.DB.GetBuild(ctx, id)
}

func (b *Builds) List(ctx context.Context, filter sdkservices.ListBuildsFilter) ([]sdktypes.Build, error) {
	if err := authz.CheckContext(
		ctx,
		sdktypes.InvalidBuildID,
		"read:list",
		authz.WithData("filter", filter),
		authz.WithAssociationWithID("project", filter.ProjectID),
	); err != nil {
		return nil, err
	}

	return b.DB.ListBuilds(ctx, filter)
}

// Download implements sdkservices.Builds.
func (b *Builds) Download(ctx context.Context, id sdktypes.BuildID) (io.ReadCloser, error) {
	if err := authz.CheckContext(ctx, id, "read:download", authz.WithConvertForbiddenToNotFound); err != nil {
		return nil, err
	}

	data, err := b.DB.GetBuildData(ctx, id)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (b *Builds) Delete(ctx context.Context, bid sdktypes.BuildID) error {
	if err := authz.CheckContext(ctx, bid, "delete:delete"); err != nil {
		return err
	}

	return b.DB.DeleteBuild(ctx, bid)
}

func (b *Builds) Describe(ctx context.Context, bid sdktypes.BuildID) (*sdkbuildfile.BuildFile, error) {
	if err := authz.CheckContext(ctx, bid, "read:describe", authz.WithConvertForbiddenToNotFound); err != nil {
		return nil, err
	}

	data, err := b.DB.GetBuildData(ctx, bid)
	if err != nil {
		return nil, err
	}

	bf, err := sdkbuildfile.Read(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}

	bf.OmitContent()

	return bf, nil
}
