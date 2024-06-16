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
	Zap      zap.Config `koanf:"zap"`
	Level    int        `koanf:"level"`    // -1 = Debug, 0 = Info, etc.
	Encoding string     `koanf:"encoding"` // "json" or "console".
}

var Configs = configset.Set[Config]{
	Default: &Config{Zap: zap.NewProductionConfig()},
	Dev:     &Config{Zap: zap.NewDevelopmentConfig()},
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
	cfg.Zap.Level.SetLevel(zapcore.Level(cfg.Level)) // Default = 0 = Info.

	if cfg.Encoding != "" {
		cfg.Zap.Encoding = cfg.Encoding
	}
	if cfg.Zap.Encoding == "console" {
		cfg.Zap.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		cfg.Zap.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)
		cfg.Zap.EncoderConfig.ConsoleSeparator = " "
	}

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
