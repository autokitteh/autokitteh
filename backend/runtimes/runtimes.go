package runtimes

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	configruntimesvc "go.autokitteh.dev/autokitteh/runtimes/configrt/runtimesvc"
	starlarkruntimesvc "go.autokitteh.dev/autokitteh/runtimes/starlarkrt/runtimesvc"
	"go.autokitteh.dev/autokitteh/sdk/sdkruntimes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

func New() sdkservices.Runtimes {
	return kittehs.Must1(sdkruntimes.New([]*sdkruntimes.Runtime{
		starlarkruntimesvc.Runtime,
		configruntimesvc.Runtime,
	}))
}
