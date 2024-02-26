package runtime

import (
	"context"
	"io/fs"

	"go.autokitteh.dev/autokitteh/runtimes/pythonrt/runtime"
	"go.autokitteh.dev/autokitteh/sdk/sdkruntimes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var Runtime = &sdkruntimes.Runtime{
	Desc: desc,
	New:  func() (sdkservices.Runtime, error) { return New(), nil },
}

type svc struct{}

func New() sdkservices.Runtime { return svc{} }

func (svc) Get() sdktypes.Runtime { return desc }

// TODO: we might want to stream the build product data as it might be big? Or we just limit the build size.
func (svc) Build(ctx context.Context, fs fs.FS, path string, values []sdktypes.Symbol) (sdktypes.BuildArtifact, error) {
	return runtime.Build(ctx, fs, path, values)
}

func (svc) Run(
	ctx context.Context,
	runID sdktypes.RunID,
	mainPath string,
	compiled map[string][]byte,
	values map[string]sdktypes.Value,
	cbs *sdkservices.RunCallbacks,
) (sdkservices.Run, error) {
	return runtime.Run(ctx, runID, mainPath, compiled, values, cbs)
}
