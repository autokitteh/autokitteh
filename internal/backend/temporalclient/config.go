package temporalclient

import (
	"path/filepath"
	"time"

	"go.temporal.io/sdk/testsuite"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/xdg"
)

type tlsConfig struct {
	Enabled      bool   `koanf:"enabled"`
	CertFilePath string `koanf:"cert_file_path"`
	KeyFilePath  string `koanf:"key_file_path"`
	Certificate  string `koanf:"certificate"`
	Key          string `koanf:"key"`
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
	DevServer                   testsuite.DevServerOptions `koanf:"dev_server"`
	DevServerStartMaxAttempts   int                        `koanf:"dev_server_start_max_attempts"`
	DevServerStartRetryInterval time.Duration              `koanf:"dev_server_start_retry_interval"`
	DevServerStartTimeout       time.Duration              `koanf:"dev_server_start_timeout"`

	TLS tlsConfig `koanf:"tls"`

	DataConverter DataConverterConfig `koanf:"data_converter"`

	EnableHelperRedirect bool `koanf:"enable_helper_redirect"`
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
			DataConverter: DataConverterConfig{
				Compress: true,
				Encryption: DataConverterEncryptionConfig{
					Encrypt: true,
				},
			},
		},
		Dev: &Config{
			Monitor:               defaultMonitorConfig,
			StartDevServerIfNotUp: true,
			DevServer: testsuite.DevServerOptions{
				LogLevel:   zapcore.WarnLevel.String(),
				EnableUI:   true,
				DBFilename: filepath.Join(xdg.DataHomeDir(), "temporal_dev.sqlite"),
			},
			DevServerStartMaxAttempts: 1,
			DevServerStartTimeout:     time.Second * 15,
			EnableHelperRedirect:      true,
		},
		Test: &Config{
			Monitor:              defaultMonitorConfig,
			AlwaysStartDevServer: true,
			DevServer: testsuite.DevServerOptions{
				LogLevel: zapcore.WarnLevel.String(),
				EnableUI: true,
			},
			DevServerStartMaxAttempts:   3,
			DevServerStartRetryInterval: time.Second,
			DevServerStartTimeout:       time.Second * 15,
		},
	}
)
