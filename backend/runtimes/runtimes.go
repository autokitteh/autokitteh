package runtimes

import (
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	configruntimesvc "go.autokitteh.dev/autokitteh/runtimes/configrt/runtimesvc"
	"go.autokitteh.dev/autokitteh/runtimes/remotert"
	starlarkruntimesvc "go.autokitteh.dev/autokitteh/runtimes/starlarkrt/runtimesvc"
	"go.autokitteh.dev/autokitteh/sdk/sdkruntimes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

type Config struct {
	RemoteRunnerEndpoints []string `koanf:"remote_runner_endpoints"`
	RemoteRunner          bool     `koanf:"enable_remote_runner"`
}

var Configs = configset.Set[Config]{
	Default: &Config{},
	Dev:     &Config{RemoteRunner: false},
}

func New(cfg Config) sdkservices.Runtimes {
	runtimes := []*sdkruntimes.Runtime{
		starlarkruntimesvc.Runtime,
		configruntimesvc.Runtime,
		// pythonrt.Runtime,
	}

	if cfg.RemoteRunner {

	}
	err := remotert.Configure(remotert.RemoteRuntimeConfig{
		ManagerAddress: []string{"localhost:9291"},
	})

	if err == nil {
		runtimes = append(runtimes, remotert.Runtime)
	} else {
		fmt.Println("failed to configure remote rt", err) //TODO: use log ?
	}

	return kittehs.Must1(sdkruntimes.New(runtimes))
}
