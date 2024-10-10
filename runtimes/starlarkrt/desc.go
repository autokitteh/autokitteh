package starlarkrt

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/runtimes/starlarkrt/runtime"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const runtimeName = "starlark"

var desc = kittehs.Must1(sdktypes.StrictRuntimeFromProto(&sdktypes.RuntimePB{
	Name: runtimeName,
	// TODO: Not sure if kitteh.star or star.kitteh.
	//       Maybe just kitteh?
	FileExtensions: runtime.Extensions,
}))
