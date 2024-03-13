package temporalclient

import (
	"time"

	"go.temporal.io/sdk/testsuite"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
)

type tlsConfig struct {
	Enabled      bool   `koanf:"enabled"`
	CertFilePath string `koanf:"cert_file_path"`
	KeyFilePath  string `koanf:"key_file_path"`
}

type MonitorConfig struct {
	CheckHealthInterval time.Duration   `koanf:"check_health_interval"`
	CheckHealthTimeout  time.Duration   `koanf:"check_health_timeout"`
	LogLevel            zap.AtomicLevel `koanf:"log_level"`
}

type Config struct {
	Monitor MonitorConfig `koanf:"monitor"`

	AlwaysStartDevServer  bool   `koanf:"always_start_dev_server"`
	StartDevServerIfNotUp bool   `koanf:"start_dev_server_if_not_up"`
	HostPort              string `koanf:"hostport"`
	Namespace             string `koanf:"namespace"`

	// DevServer.ClientOptions is not used.
	DevServer testsuite.DevServerOptions `koanf:"dev_server"`
	TLS       tlsConfig                  `koanf:"tls"`
}

var (
	defaultMonitorConfig = MonitorConfig{
		CheckHealthInterval: time.Minute,
		CheckHealthTimeout:  10 * time.Second,
		LogLevel:            zap.NewAtomicLevelAt(zapcore.WarnLevel),
	}

	Configs = configset.Set[Config]{
		Default: &Config{
			Monitor: defaultMonitorConfig,
		},
		Dev: &Config{
			Monitor:               defaultMonitorConfig,
			StartDevServerIfNotUp: true,
			DevServer: testsuite.DevServerOptions{
				LogLevel: zapcore.WarnLevel.String(),
			},
		},
		Test: &Config{
			Monitor:              defaultMonitorConfig,
			AlwaysStartDevServer: true,
			DevServer: testsuite.DevServerOptions{
				LogLevel: zapcore.WarnLevel.String(),
			},
		},
	}
)
