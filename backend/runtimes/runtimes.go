package runtimes

import (
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	configruntimesvc "go.autokitteh.dev/autokitteh/runtimes/configrt/runtimesvc"
	pythonruntime "go.autokitteh.dev/autokitteh/runtimes/pythonrt"
	starlarkruntimesvc "go.autokitteh.dev/autokitteh/runtimes/starlarkrt/runtimesvc"
	"go.autokitteh.dev/autokitteh/sdk/sdkruntimes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.uber.org/zap"
)

type Config struct {
	RemoteRunnerEndpoints []string `koanf:"remote_runner_endpoints"`
	RemoteRunner          bool     `koanf:"enable_remote_runner"`
	WorkerAddress         string   `koanf:"worker_address"`
	// TODO: This is a hack to prevent running configure on pythonrt in each test
	// which currently install venv everytime and takes a really long time
	// need to find a way to share the venv once for all tests
	LazyLoadLocalVEnv bool `koanf:"lazy_load_local_venv"`
}

var Configs = configset.Set[Config]{
	Default: &Config{
		RemoteRunnerEndpoints: []string{"localhost:9291"},
		RemoteRunner:          false,
		WorkerAddress:         "localhost:9980",
	},
	Test: &Config{
		LazyLoadLocalVEnv: true,
	},
}

func New(cfg *Config, l *zap.Logger) sdkservices.Runtimes {
	runtimes := []*sdkruntimes.Runtime{
		starlarkruntimesvc.Runtime,
		configruntimesvc.Runtime,
		pythonruntime.Runtime,
	}

	//TODO: need to rethink
	if cfg == nil {
		cfg = Configs.Default
	}

	if cfg.RemoteRunner {
		if len(cfg.RemoteRunnerEndpoints) == 0 {
			panic("remote runner is enabled but no runner endpoints provided")
		}
		if err := pythonruntime.ConfigureRemoteRunnerManager(pythonruntime.RemoteRuntimeConfig{
			ManagerAddress: cfg.RemoteRunnerEndpoints,
			WorkerAddress:  cfg.WorkerAddress,
		}); err != nil {
			l.Panic("configure remote runner manager", zap.Error(err))
		}
		l.Info("remote runner configued")
	} else {
		if err := pythonruntime.ConfigureLocalRunnerManager(l,
			pythonruntime.LocalRunnerConfig{
				WorkerAddress: cfg.WorkerAddress,
				LazyLoadVEnv:  cfg.LazyLoadLocalVEnv,
			},
		); err != nil {
			l.Panic("configure local runner manager", zap.Error(err))
		}

		l.Info("local runner configured", zap.String("worker address", cfg.WorkerAddress))
	}

	pythonruntime.ConfigureWorkerGRPCHandler(l)

	return kittehs.Must1(sdkruntimes.New(runtimes))
}
