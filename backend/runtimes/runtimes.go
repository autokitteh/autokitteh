package runtimes

import (
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	configruntimesvc "go.autokitteh.dev/autokitteh/runtimes/configrt/runtimesvc"
	"go.autokitteh.dev/autokitteh/runtimes/pythonrt"
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
	Default: &Config{
		RemoteRunnerEndpoints: []string{"localhost:9291"},
	},
	Dev: &Config{RemoteRunner: false},
}

func New(cfg *Config) sdkservices.Runtimes {
	runtimes := []*sdkruntimes.Runtime{
		starlarkruntimesvc.Runtime,
		configruntimesvc.Runtime,
	}

	if cfg.RemoteRunner {
		if len(cfg.RemoteRunnerEndpoints) == 0 {
			panic("remote runner is enabled but no runner endpoints provided")
		}
		if err := remotert.Configure(remotert.RemoteRuntimeConfig{
			ManagerAddress: cfg.RemoteRunnerEndpoints,
		}); err != nil {
			fmt.Printf("Could not start remote RT %s", err)
			panic("")
		}
		fmt.Println("Remote runtime configured")
		runtimes = append(runtimes, remotert.Runtime)
	} else {
		fmt.Println("Local runtime configued")
		runtimes = append(runtimes, pythonrt.Runtime)
	}

	return kittehs.Must1(sdkruntimes.New(runtimes))
}
