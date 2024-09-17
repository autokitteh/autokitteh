package runtimes

import (
	"errors"
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/httpsvc"
	configruntimesvc "go.autokitteh.dev/autokitteh/runtimes/configrt/runtimesvc"
	pythonruntime "go.autokitteh.dev/autokitteh/runtimes/pythonrt"
	starlarkruntimesvc "go.autokitteh.dev/autokitteh/runtimes/starlarkrt/runtimesvc"
	"go.autokitteh.dev/autokitteh/sdk/sdkruntimes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.uber.org/zap"
)

type Config struct {
	RemoteRunnerEndpoints []string `koanf:"remote_runner_endpoints"`
	//TODO: maybe should be runner type which can be local/docker/remote/ ?
	EnableRemoteRunner bool   `koanf:"enable_remote_runner"`
	WorkerAddress      string `koanf:"worker_address"`
	// TODO: This is a hack to prevent running configure on pythonrt in each test
	// which currently install venv everytime and takes a really long time
	// need to find a way to share the venv once for all tests
	LazyLoadLocalVEnv bool `koanf:"lazy_load_local_venv"`
}

var Configs = configset.Set[Config]{
	Default: &Config{
		RemoteRunnerEndpoints: []string{"localhost:9291"},
		EnableRemoteRunner:    false,
		// WorkerAddress:         "localhost:9980",
	},
	Test: &Config{
		LazyLoadLocalVEnv: true,
	},
}

func New(cfg *Config, l *zap.Logger, svc httpsvc.Svc) (sdkservices.Runtimes, error) {
	runtimes := []*sdkruntimes.Runtime{
		starlarkruntimesvc.Runtime,
		configruntimesvc.Runtime,
		pythonruntime.Runtime,
	}

	//TODO: need to rethink
	if cfg == nil {
		cfg = Configs.Default
	}

	if cfg.EnableRemoteRunner {
		if len(cfg.RemoteRunnerEndpoints) == 0 {
			return nil, errors.New("remote runner is enabled but no runner endpoints provided")
		}
		if err := pythonruntime.ConfigureRemoteRunnerManager(pythonruntime.RemoteRuntimeConfig{
			ManagerAddress: cfg.RemoteRunnerEndpoints,
			WorkerAddress:  cfg.WorkerAddress,
		}); err != nil {
			return nil, fmt.Errorf("configure remote runner manager: %w", err)
		}
		l.Info("remote runner configued")
	} else {
		if err := pythonruntime.ConfigureLocalRunnerManager(l,
			pythonruntime.LocalRunnerConfig{
				WorkerAddress:         cfg.WorkerAddress,
				LazyLoadVEnv:          cfg.LazyLoadLocalVEnv,
				WorkerAddressProvider: func() string { return svc.Addr() },
			},
		); err != nil {
			return nil, fmt.Errorf("configure local runner manager: %w", err)
		}

		l.Info("local runner configured")
	}

	return sdkruntimes.New(runtimes)
}
