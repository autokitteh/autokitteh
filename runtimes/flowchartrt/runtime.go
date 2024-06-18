package flowchartrt

import (
	"strings"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkruntimes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	exts = []string{"flowchart.yaml", "flowchart.json"}

	desc = kittehs.Must1(sdktypes.StrictRuntimeFromProto(&sdktypes.RuntimePB{
		Name:           "flowchart",
		FileExtensions: exts,
	}))

	Runtime = &sdkruntimes.Runtime{
		Desc: desc,
		New:  func() (sdkservices.Runtime, error) { return New(), nil },
	}
)

func isFlowchartPath(path string) bool {
	for _, ext := range exts {
		if strings.HasSuffix(path, "."+ext) {
			return true
		}
	}

	return false
}

type rt struct{}

func New() sdkservices.Runtime { return rt{} }

func (rt) Get() sdktypes.Runtime { return desc }
