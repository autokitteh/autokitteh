package logger

import (
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdklogger"
)

type Config struct {
	Zap      zap.Config `koanf:"zap"`
	Level    string     `koanf:"level"`    // Case-insensitive, default = info.
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
	// Optional override for the default level (0 = info).
	if cfg.Level != "" {
		// Accept case-insensitive names, not just lower-case or all-caps.
		level, err := zapcore.ParseLevel(strings.ToLower(cfg.Level))
		if err != nil {
			// Temporary: numeric level IDs, as used internally by Zap.
			if n, e := strconv.Atoi(cfg.Level); e == nil {
				level = zapcore.Level(n)
			} else {
				return nil, err
			}
		}
		cfg.Zap.Level.SetLevel(level)
	}

	// Optional override for the default encoding:
	// AK default mode = Zap production config = "json" encoding,
	// AK dev mode = Zap development config = "console" encoding.
	if cfg.Encoding != "" {
		cfg.Zap.Encoding = cfg.Encoding
	}

	// Regardless of AK mode, if the encoding is "console" (whether
	// as a default or as an override), tweak it for easier reading.
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
