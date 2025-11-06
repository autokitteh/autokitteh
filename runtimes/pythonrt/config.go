package pythonrt

import (
	"time"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
)

type Config struct {
	RemoteRunnerEndpoints []string `koanf:"remote_runner_endpoints"`
	WorkerAddress         string   `koanf:"worker_address"`
	// TODO: This is a hack to prevent running configure on pythonrt in each test
	// which currently install venv everytime and takes a really long time
	// need to find a way to share the venv once for all tests
	LazyLoadLocalVEnv bool `koanf:"lazy_load_local_venv"`
	LogRunnerCode     bool `koanf:"log_runner_code"`
	LogBuildCode      bool `koanf:"log_build_code"`

	RunnerType string `koanf:"runner_type"`

	DelayedStartPrintTimeout time.Duration `koanf:"delayed_start_print_timeout"`

	MaxMemoryPerWorkflowMB int64   `koanf:"max_memory_per_workflow_mb"`
	MaxCPUsPerWorkflow     float32 `koanf:"max_cpus_per_workflow"`

	// Support simultaneous multiple venvs for local runner.
	LocalMultiVenv bool `koanf:"local_multi_venv"`
}

var Configs = configset.Set[Config]{
	Default: &Config{
		RunnerType:               "local",
		DelayedStartPrintTimeout: 10 * time.Second,
		MaxMemoryPerWorkflowMB:   512, // 512MB
		MaxCPUsPerWorkflow:       1,
	},
	Dev: &Config{
		LogRunnerCode:            true,
		LogBuildCode:             true,
		DelayedStartPrintTimeout: 10 * time.Second,
		// This is experimental, so enable by default only in dev.
		LocalMultiVenv: true,
	},
	Test: &Config{
		LazyLoadLocalVEnv:        true,
		DelayedStartPrintTimeout: 0,
	},
}
