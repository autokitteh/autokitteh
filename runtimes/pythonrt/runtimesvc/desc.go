package runtime

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/runtimes/pythonrt/runtime"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const runtimeName = "python"

var desc = kittehs.Must1(sdktypes.StrictRuntimeFromProto(&sdktypes.RuntimePB{
	Name:           runtimeName,
	FileExtensions: runtime.Extensions,
}))
