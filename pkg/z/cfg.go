package z

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Zap logger configuration.
type Config struct {
	// Minimal log level to be emitted.
	Level string `envconfig:"LEVEL" default:"info" json:"level"`

	// Dev determines log output format. True for human readable (appropiate
	// for console output), and False for machine readable (appropiate for
	// automated consumption as a JSON blob).
	Dev bool `envconfig:"DEV" default:"true" json:"dev"`
}

var DefaultOpts = []zap.Option{
	zap.AddCaller(),
	zap.AddStacktrace(zapcore.ErrorLevel),
}

// New returns a new Zap logger based on the given configuration.
func New(cfg Config, f func(*zap.Config), zopts []zap.Option) (*zap.SugaredLogger, error) {
	zcfg := zap.NewProductionConfig()
	if cfg.Dev {
		zcfg = zap.NewDevelopmentConfig()

		zcfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		zcfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02T15:04:05.99")
	}

	if err := zcfg.Level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return nil, fmt.Errorf(`invalid log level "%s": %w`, cfg.Level, err)
	}

	if f != nil {
		f(&zcfg)
	}

	if zopts == nil {
		zopts = DefaultOpts
	}

	z, err := zcfg.Build(zopts...)
	if err != nil {
		return nil, fmt.Errorf("failed initializing log: %w", err)
	}

	return z.Sugar(), nil
}
