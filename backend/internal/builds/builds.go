package builds

import (
	"bytes"
	"context"
	"io"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.autokitteh.dev/autokitteh/backend/internal/db"
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
	build, err := build.Update(func(pb *sdktypes.BuildPB) {
		pb.BuildId = sdktypes.NewBuildID().String()
		pb.CreatedAt = timestamppb.New(time.Now())
	})
	if err != nil {
		return nil, err
	}

	if err := b.DB.SaveBuild(ctx, build, data); err != nil {
		return nil, err
	}

	return sdktypes.GetBuildID(build), nil
}

func (b *Builds) Remove(ctx context.Context, id sdktypes.BuildID) error {
	return b.DB.DeleteBuild(ctx, id)
}
