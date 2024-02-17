package httpsvc

import (
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/backend/configset"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
)

type httpH2CConfig struct {
	Enable bool `koanf:"enable"`
}

type ngrokConfig struct {
	Enable    bool   `koanf:"enable"`
	AuthToken string `koanf:"auth_token"`
	Domain    string `koanf:"domain"`
}

type LoggerConfig struct {
	ImportantLevel    zap.AtomicLevel `koanf:"important_log_level"`
	NonimportantLevel zap.AtomicLevel `koanf:"nonimportant_log_level"`
	ErrorsLevel       zap.AtomicLevel `koanf:"trace_errors_log_level"`

	// Requests that their paths matches any one of these regexes will be
	// logged in the nonimportant log level.
	NonimportantRegexes []string `koanf:"nonimportant_regexes"`
}

type CORSConfig struct {
	AllowedOrigins   []string `koanf:"allowed_origins"`
	AllowedMethods   []string `koanf:"allowed_methods"`
	AllowedHeaders   []string `koanf:"allowed_headers"`
	AllowCredentials bool     `koanf:"allow_credentials"`
}

type Config struct {
	Addr string        `koanf:"addr"`
	H2C  httpH2CConfig `koanf:"h2c"`

	// If not empty, write HTTP port to this file.
	// This is useful when starting with port 0, which means to get
	// the next port. This is done in testing to start on an unused
	// port to avoid conflict with an already running service.
	AddrFilename         string `koanf:"addr_filename"`
	EnableGRPCReflection bool   `koanf:"reflection"`

	Ngrok ngrokConfig `koanf:"ngrok"`

	Logger LoggerConfig `koanf:"logger"`

	CORS CORSConfig `koanf:"cors"`
}

var Configs = configset.Set[Config]{
	Default: &Config{
		Addr:                 "0.0.0.0:" + sdkclient.DefaultPort,
		H2C:                  httpH2CConfig{Enable: true},
		EnableGRPCReflection: true,
		CORS: CORSConfig{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH", "HEAD"},
			AllowedHeaders:   []string{"*"},
			AllowCredentials: true,
		},
		Logger: LoggerConfig{
			NonimportantLevel: zap.NewAtomicLevelAt(zap.DebugLevel),
			ImportantLevel:    zap.NewAtomicLevelAt(zap.InfoLevel),
			ErrorsLevel:       zap.NewAtomicLevelAt(zap.WarnLevel),
			NonimportantRegexes: []string{
				`.*/List.*`,
				`.*/Get.*`,
			},
		},
	},
}
