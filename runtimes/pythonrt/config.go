package pythonrt

import "go.autokitteh.dev/autokitteh/internal/backend/configset"

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

var Configs = configset.Set[Config]{
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
