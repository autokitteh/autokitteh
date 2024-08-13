package runtimes

import (
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	configruntimesvc "go.autokitteh.dev/autokitteh/runtimes/configrt/runtimesvc"
	"go.autokitteh.dev/autokitteh/runtimes/remotert"
	starlarkruntimesvc "go.autokitteh.dev/autokitteh/runtimes/starlarkrt/runtimesvc"
	"go.autokitteh.dev/autokitteh/sdk/sdkruntimes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

func New() sdkservices.Runtimes {
	runtimes := []*sdkruntimes.Runtime{
		starlarkruntimesvc.Runtime,
		configruntimesvc.Runtime,
		// pythonruntime.Runtime,
	}
	err := remotert.Configure(remotert.RemoteRuntimeConfig{
		RunnerAddress: []string{"localhost:9291"},
	})

	if err == nil {
		runtimes = append(runtimes, remotert.Runtime)
	} else {
		fmt.Println("failed to configure remote rt", err) //TODO: use log ?
	}

	return kittehs.Must1(sdkruntimes.New(runtimes))
}
