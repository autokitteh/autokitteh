package runtimes

import (
	"errors"
	"fmt"
	"net"

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
	PythonRunnerType   string `koanf:"python_runner_type"`
	WorkerAddress      string `koanf:"worker_address"`
	// TODO: This is a hack to prevent running configure on pythonrt in each test
	// which currently install venv everytime and takes a really long time
	// need to find a way to share the venv once for all tests
	LazyLoadLocalVEnv bool `koanf:"lazy_load_local_venv"`
	LogRunnerCode     bool `koanf:"log_runner_code"`
}

var Configs = configset.Set[Config]{
	Default: &Config{
		EnableRemoteRunner: false,
	},
	Test: &Config{
		LazyLoadLocalVEnv: true,
	},
	Dev: &Config{
		LogRunnerCode:    true,
		PythonRunnerType: "docker",
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

	switch cfg.PythonRunnerType {
	case "docker":
		if err := pythonruntime.ConfigureDockerRunnerManager(l, pythonruntime.DockerRuntimeConfig{
			WorkerAddressProvider: func() string {
				_, port, _ := net.SplitHostPort(svc.Addr())
				return fmt.Sprintf("host.docker.internal:%s", port)
			},
		}); err != nil {
			return nil, fmt.Errorf("configure docker runner manager: %w", err)
		}
		l.Info("docker runner configured")
	case "remote":
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
	default:
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
