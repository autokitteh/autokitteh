package builds

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/sdk/sdkbuildfile"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Builds struct {
	fx.In

	Z  *zap.Logger
	DB db.DB
}

func New(b Builds) sdkservices.Builds { return &b }

func (b *Builds) Get(ctx context.Context, id sdktypes.BuildID) (sdktypes.Build, error) {
	return sdkerrors.IgnoreNotFoundErr(b.DB.GetBuild(ctx, id))
}

func (b *Builds) List(ctx context.Context, filter sdkservices.ListBuildsFilter) ([]sdktypes.Build, error) {
	return b.DB.ListBuilds(ctx, filter)
}

// Download implements sdkservices.Builds.
func (b *Builds) Download(ctx context.Context, id sdktypes.BuildID) (io.ReadCloser, error) {
	data, err := b.DB.GetBuildData(ctx, id)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (b *Builds) Save(ctx context.Context, build sdktypes.Build, data []byte) (sdktypes.BuildID, error) {
	// make sure this at least tries to pretend to be a build file.
	if _, err := sdkbuildfile.ReadVersion(bytes.NewReader(data)); err != nil {
		return sdktypes.InvalidBuildID, fmt.Errorf("read version: %w", err)
	}

	build = build.
		WithNewID().
		WithCreatedAt(time.Now())

	if err := b.DB.SaveBuild(ctx, build, data); err != nil {
		return sdktypes.InvalidBuildID, err
	}

	return build.ID(), nil
}

func (b *Builds) Delete(ctx context.Context, id sdktypes.BuildID) error {
	return b.DB.DeleteBuild(ctx, id)
}

func (b *Builds) Describe(ctx context.Context, bid sdktypes.BuildID) (*sdkbuildfile.BuildFile, error) {
	r, err := b.Download(ctx, bid)
	if err != nil {
		return nil, fmt.Errorf("download: %w", err)
	}
	defer r.Close()

	bf, err := sdkbuildfile.Read(r)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}

	bf.OmitContent()

	return bf, nil
}
