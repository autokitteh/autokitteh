package pythonrt

import "go.autokitteh.dev/autokitteh/internal/backend/configset"

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
}

var Configs = configset.Set[Config]{
	Default: &Config{
		RunnerType: "local",
	},
	Test: &Config{
		LazyLoadLocalVEnv: true,
	},
	Dev: &Config{
		LogRunnerCode: true,
		LogBuildCode:  true,
	},
}
