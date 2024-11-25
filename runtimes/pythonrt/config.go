package pythonrt

import "go.autokitteh.dev/autokitteh/internal/backend/configset"

type Config struct {
	// Comma separated list of remote runner endpoints
	RemoteRunnerEndpoints string `koanf:"remote_runner_endpoints"`
	// Self IP address of the server running the workflows (temporal worker)
	WorkerAddress string `koanf:"worker_address"`
	// TODO: This is a hack to prevent running configure on pythonrt in each test
	// which currently install venv everytime and takes a really long time
	// need to find a way to share the venv once for all tests
	LazyLoadLocalVEnv bool `koanf:"lazy_load_local_venv"`
	LogRunnerCode     bool `koanf:"log_runner_code"`
	LogBuildCode      bool `koanf:"log_build_code"`

	// Currently local, docker, remote. local is default.
	// remote requires setting remote runner endpoints as well
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
