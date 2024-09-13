package runtimes

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/runtimes/configrt"
	"go.autokitteh.dev/autokitteh/runtimes/pythonrt"
	"go.autokitteh.dev/autokitteh/runtimes/starlarkrt"
	"go.autokitteh.dev/autokitteh/sdk/sdkruntimes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

func New() sdkservices.Runtimes {
	return kittehs.Must1(sdkruntimes.New([]*sdkruntimes.Runtime{
		starlarkrt.New(),
		configrt.New(),
		pythonrt.New(),
	}))
}
