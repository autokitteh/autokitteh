package sdkservices

import (
	"context"
	"io"

	"go.autokitteh.dev/autokitteh/sdk/sdkbuildfile"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type ListBuildsFilter struct {
	Limit uint32
}

type Builds interface {
	Get(ctx context.Context, id sdktypes.BuildID) (sdktypes.Build, error)
	List(ctx context.Context, filter ListBuildsFilter) ([]sdktypes.Build, error)
	Download(ctx context.Context, id sdktypes.BuildID) (io.ReadCloser, error)
	Save(ctx context.Context, build sdktypes.Build, data []byte) (sdktypes.BuildID, error)
	Delete(ctx context.Context, id sdktypes.BuildID) error
	Describe(ctx context.Context, id sdktypes.BuildID) (*sdkbuildfile.BuildFile, error)
}
