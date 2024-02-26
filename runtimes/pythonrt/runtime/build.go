package runtime

import (
	"context"
	"io/fs"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func Build(ctx context.Context, fs fs.FS, path string, symbols []sdktypes.Symbol) (sdktypes.BuildArtifact, error) {
	return nil, sdkerrors.ErrNotImplemented
}
