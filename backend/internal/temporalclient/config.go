package temporalclient

import (
	"time"

	"go.temporal.io/sdk/testsuite"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go.autokitteh.dev/autokitteh/backend/configset"
)

type tlsConfig struct {
	Enabled      bool   `koanf:"enabled"`
	CertFilePath string `koanf:"cert_file_path"`
	KeyFilePath  string `koanf:"key_file_path"`
}

type Config struct {
	AlwaysStartDevServer  bool            `koanf:"always_start_dev_server"`
	StartDevServerIfNotUp bool            `koanf:"start_dev_server_if_not_up"`
	HostPort              string          `koanf:"hostport"`
	Namespace             string          `koanf:"namespace"`
	CheckHealthInterval   time.Duration   `koanf:"check_health_interval"`
	CheckHealthTimeout    time.Duration   `koanf:"check_health_timeout"`
	LogLevel              zap.AtomicLevel `koanf:"log_level"`

	// DevServer.ClientOptions is not used.
	DevServer testsuite.DevServerOptions `koanf:"dev_server"`
	TLS       tlsConfig                  `koanf:"tls"`
}

var Configs = configset.Set[Config]{
	Default: &Config{
		CheckHealthInterval: time.Minute,
		CheckHealthTimeout:  10 * time.Second,
		LogLevel:            zap.NewAtomicLevelAt(zapcore.WarnLevel),
	},
	Dev: &Config{
		CheckHealthInterval: time.Minute,
		CheckHealthTimeout:  10 * time.Second,
		LogLevel:            zap.NewAtomicLevelAt(zapcore.WarnLevel),

		StartDevServerIfNotUp: true,
		DevServer: testsuite.DevServerOptions{
			LogLevel: zapcore.WarnLevel.String(),
		},
	},
	Test: &Config{
		CheckHealthInterval: time.Minute,
		CheckHealthTimeout:  10 * time.Second,
		LogLevel:            zap.NewAtomicLevelAt(zapcore.WarnLevel),

		AlwaysStartDevServer: true,
		DevServer: testsuite.DevServerOptions{
			LogLevel: zapcore.WarnLevel.String(),
		},
	},
}
