package sdkruntimesclient

import (
	"context"
	"io/fs"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type runtime struct {
	desc sdktypes.Runtime
}

func (r *runtime) Get() sdktypes.Runtime { return r.desc }

func (r *runtime) Build(context.Context, fs.FS, string, []sdktypes.Symbol) (sdktypes.BuildArtifact, error) {
	return sdktypes.InvalidBuildArtifact, sdkerrors.ErrNotImplemented
}

func (r *runtime) Run(context.Context, sdktypes.RunID, sdktypes.SessionID, string, map[string][]byte, map[string]sdktypes.Value, sdkservices.RunCallbacks) (sdkservices.Run, error) {
	return nil, sdkerrors.ErrNotImplemented
}
