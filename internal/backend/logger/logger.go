package logger

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdklogger"
)

type Config struct {
	Zap zap.Config `koanf:"zap"`
}

var defaultZapLevel = zap.NewAtomicLevelAt(zap.DebugLevel)

var Configs = configset.Set[Config]{
	Default: &Config{Zap: zap.NewProductionConfig()},
	Dev: &Config{
		Zap: (func() (cfg zap.Config) {
			cfg = zap.NewDevelopmentConfig()
			cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
			cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)
			cfg.EncoderConfig.ConsoleSeparator = " "
			cfg.Level = defaultZapLevel
			return
		})(),
	},
}

func (c *Config) WithDebug(debug bool) *Config {
	cc := *c
	c = &cc

	if debug {
		c.Zap.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	} else {
		c.Zap.Level = defaultZapLevel
	}

	return c
}

type onFatalHook struct{}

func (onFatalHook) OnWrite(ce *zapcore.CheckedEntry, fs []zapcore.Field) {
	// This is a useful place for a breakpoint to catch all fatals.
	zapcore.WriteThenGoexit.OnWrite(ce, fs)
}

type onPanicHook struct{}

func (onPanicHook) OnWrite(ce *zapcore.CheckedEntry, fs []zapcore.Field) {
	// This is a useful place for a breakpoint to catch all panics.
	zapcore.WriteThenPanic.OnWrite(ce, fs)
}

func New(cfg *Config) (*zap.Logger, error) {
	z, err := cfg.Zap.Build(
		zap.WithFatalHook(onFatalHook{}),
		zap.WithPanicHook(onPanicHook{}),
		zap.AddStacktrace(zap.ErrorLevel),
	)
	if err != nil {
		return nil, err
	}

	zap.ReplaceGlobals(z)

	sdklogger.SetGlobalLogger(z.Named("sdk").Sugar())

	kittehs.SetPanicFunc(func(msg any) { z.Panic("panic", zap.Any("msg", msg)) })

	return z, nil
}
