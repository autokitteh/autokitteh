package configrt

import (
	"context"
	"io/fs"

	"go.autokitteh.dev/autokitteh/runtimes/configrt/runtime"
	"go.autokitteh.dev/autokitteh/sdk/sdkruntimes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type svc struct{}

func New() *sdkruntimes.Runtime {
	return &sdkruntimes.Runtime{
		Desc: desc,
		New:  func() (sdkservices.Runtime, error) { return &svc{}, nil },
	}
}

func (svc) Get() sdktypes.Runtime { return desc }

func (svc) Build(ctx context.Context, fs fs.FS, path string, _ []sdktypes.Symbol) (sdktypes.BuildArtifact, error) {
	return runtime.Build(ctx, fs, path)
}

func (svc) Run(
	_ context.Context,
	runID sdktypes.RunID,
	sessionID sdktypes.SessionID,
	mainPath string,
	compiled map[string][]byte,
	_ map[string]sdktypes.Value,
	_ sdkservices.RunCallbacks,
) (sdkservices.Run, error) {
	return runtime.Run(runID, mainPath, compiled)
}
