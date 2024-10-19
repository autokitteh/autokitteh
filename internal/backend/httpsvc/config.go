package httpsvc

import (
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
)

type httpH2CConfig struct {
	Enable bool `koanf:"enable"`
}

type LoggerConfig struct {
	ImportantLevel   zap.AtomicLevel `koanf:"important_log_level"`
	UnimportantLevel zap.AtomicLevel `koanf:"unimportant_log_level"`
	ErrorsLevel      zap.AtomicLevel `koanf:"trace_errors_log_level"`

	// URL requests with paths that match any one of these regular
	// expressions will be logged in the nonimportant log level.
	UnimportantRegexes []string `koanf:"nonimportant_regexes"`

	// URL requests with paths that match any one of these
	// regular expressions will not be logged at all.
	UnloggedRegexes []string `koanf:"unlogged_regexes"`
}

type CORSConfig struct {
	AllowedOrigins   []string `koanf:"allowed_origins"`
	AllowedMethods   []string `koanf:"allowed_methods"`
	AllowedHeaders   []string `koanf:"allowed_headers"`
	AllowCredentials bool     `koanf:"allow_credentials"`
}

type Config struct {
	// local server address, set to run server on different port
	Addr    string `koanf:"addr"`
	AuxAddr string `koanf:"aux_addr"`

	// ak service url, used in client to connect to connect to specific ak server
	ServiceURL string `koanf:"service_url"`

	H2C httpH2CConfig `koanf:"h2c"`

	// If not empty, write main HTTP port to this file.
	// This is useful when starting with port 0, which means to get
	// the next port. This is done in testing to start on an unused
	// port to avoid conflict with an already running service.
	AddrFilename string `koanf:"addr_filename"`

	// Enable gRPC reflection for the main mux.
	EnableGRPCReflection bool `koanf:"reflection"`

	Logger LoggerConfig `koanf:"logger"`

	CORS CORSConfig `koanf:"cors"`
}

var defaultConfg = Config{
	Addr:                 "0.0.0.0:" + sdkclient.DefaultPort,
	AuxAddr:              "0.0.0.0:9983",
	ServiceURL:           sdkclient.DefaultCloudURL,
	H2C:                  httpH2CConfig{Enable: true},
	EnableGRPCReflection: true,
	CORS: CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH", "HEAD"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	},
	Logger: LoggerConfig{
		UnimportantLevel: zap.NewAtomicLevelAt(zap.DebugLevel),
		ImportantLevel:   zap.NewAtomicLevelAt(zap.InfoLevel),
		ErrorsLevel:      zap.NewAtomicLevelAt(zap.WarnLevel),
		UnimportantRegexes: []string{
			`^/autokitteh.+/(Get|List).*`, // gRPC Get and List methods
			`^/oauth/|/oauth$|/save$`,     // Connection initialization
		},
		UnloggedRegexes: []string{
			`/(healthz|readyz)$`,                           // Kubernetes health checks
			`\.(css|html|ico|js|png|svg|txt|webmanifest)$`, // Static web content
		},
	},
}

var Configs = configset.Set[Config]{
	Default: &defaultConfg,
	Dev: func() *Config {
		c := defaultConfg
		c.ServiceURL = sdkclient.DefaultLocalURL
		return &c
	}(),
}
