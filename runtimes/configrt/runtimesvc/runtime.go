package runtime

import (
	"context"
	"net/url"

	"go.autokitteh.dev/autokitteh/runtimes/configrt/runtime"
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

func (svc) Build(ctx context.Context, rootURL *url.URL, path string, _ []sdktypes.Symbol) (sdktypes.BuildArtifact, error) {
	return runtime.Build(ctx, rootURL, path)
}

func (svc) Run(
	_ context.Context,
	runID sdktypes.RunID,
	mainPath string,
	compiled map[string][]byte,
	_ map[string]sdktypes.Value,
	_ *sdkservices.RunCallbacks,
) (sdkservices.Run, error) {
	return runtime.Run(runID, mainPath, compiled)
}
