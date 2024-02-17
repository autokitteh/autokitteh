package runtime

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/runtimes/configrt/parsers"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const runtimeName = "config"

var desc = kittehs.Must1(sdktypes.StrictRuntimeFromProto(&sdktypes.RuntimePB{
	Name:           runtimeName,
	FileExtensions: parsers.Extensions,
}))
