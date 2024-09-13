package starlarkrt

import (
	"context"
	"io/fs"

	"go.autokitteh.dev/autokitteh/runtimes/starlarkrt/runtime"
	"go.autokitteh.dev/autokitteh/sdk/sdkruntimes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Config = runtime.Config

var Configs = runtime.Configs

func New() *sdkruntimes.Runtime { return NewFromConfig(runtime.Configs.Default) }

func NewFromConfig(cfg *Config) *sdkruntimes.Runtime {
	return &sdkruntimes.Runtime{
		Desc: desc,
		New:  func() (sdkservices.Runtime, error) { return svc{cfg: cfg}, nil },
	}
}

type svc struct{ cfg *Config }

func (svc) Get() sdktypes.Runtime { return desc }

// TODO: we might want to stream the build product data as it might be big? Or we just limit the build size.
func (s svc) Build(ctx context.Context, fs fs.FS, path string, values []sdktypes.Symbol) (sdktypes.BuildArtifact, error) {
	return runtime.Build(ctx, s.cfg, fs, path, values)
}

func (s svc) Run(
	ctx context.Context,
	runID sdktypes.RunID,
	mainPath string,
	compiled map[string][]byte,
	values map[string]sdktypes.Value,
	cbs *sdkservices.RunCallbacks,
) (sdkservices.Run, error) {
	return runtime.Run(ctx, s.cfg, runID, mainPath, compiled, values, cbs)
}
