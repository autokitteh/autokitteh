package runtimes

import (
	"errors"
	"fmt"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/config"
	"go.autokitteh.dev/autokitteh/internal/backend/httpsvc"
	configruntimesvc "go.autokitteh.dev/autokitteh/runtimes/configrt/runtimesvc"
	pythonruntime "go.autokitteh.dev/autokitteh/runtimes/pythonrt"
	starlarkruntimesvc "go.autokitteh.dev/autokitteh/runtimes/starlarkrt/runtimesvc"
	"go.autokitteh.dev/autokitteh/sdk/sdkruntimes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

type Config struct {
	RemoteRunnerEndpoints []string `koanf:"remote_runner_endpoints"`
	// TODO: maybe should be runner type which can be local/docker/remote/ ?
	EnableRemoteRunner bool   `koanf:"enable_remote_runner"`
	WorkerAddress      string `koanf:"worker_address"`
	// TODO: This is a hack to prevent running configure on pythonrt in each test
	// which currently install venv everytime and takes a really long time
	// need to find a way to share the venv once for all tests
	LazyLoadLocalVEnv bool `koanf:"lazy_load_local_venv"`
	LogRunnerCode     bool `koanf:"log_runner_code"`
}

func (c Config) Validate() error {
	if c.EnableRemoteRunner && len(c.RemoteRunnerEndpoints) == 0 {
		return errors.New("remote runner is enabled but no runner endpoints provided")
	}

	return nil
}

var Configs = config.Set[Config]{
	Default: &Config{
		EnableRemoteRunner: false,
	},
	Test: &Config{
		LazyLoadLocalVEnv: true,
	},
	Dev: &Config{
		LogRunnerCode: true,
	},
}

func New(cfg *Config, l *zap.Logger, svc httpsvc.Svc) (sdkservices.Runtimes, error) {
	runtimes := []*sdkruntimes.Runtime{
		starlarkruntimesvc.Runtime,
		configruntimesvc.Runtime,
		pythonruntime.Runtime,
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	if cfg.EnableRemoteRunner {
		if err := pythonruntime.ConfigureRemoteRunnerManager(pythonruntime.RemoteRuntimeConfig{
			ManagerAddress: cfg.RemoteRunnerEndpoints,
			WorkerAddress:  cfg.WorkerAddress,
		}); err != nil {
			return nil, fmt.Errorf("configure remote runner manager: %w", err)
		}
		l.Info("remote runner configued")
	} else {
		if err := pythonruntime.ConfigureLocalRunnerManager(l,
			pythonruntime.LocalRunnerManagerConfig{
				WorkerAddress:         cfg.WorkerAddress,
				LazyLoadVEnv:          cfg.LazyLoadLocalVEnv,
				WorkerAddressProvider: func() string { return svc.Addr() },
				LogCodeRunnerCode:     cfg.LogRunnerCode,
			},
		); err != nil {
			return nil, fmt.Errorf("configure local runner manager: %w", err)
		}

		l.Info("local runner configured")
	}

	return sdkruntimes.New(runtimes)
}
