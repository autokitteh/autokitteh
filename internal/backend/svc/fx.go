package svc

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
)

// TODO: need this so RunOptions would not needed to be passed every time Component is called. This is ugly, fix.
var fxRunOpts RunOptions

func setFXRunOpts(opts RunOptions) { fxRunOpts = opts }

func fxGetConfig[T any](path string, def T) func(c *Config) (*T, error) {
	return func(c *Config) (*T, error) {
		return GetConfig(c, path, def)
	}
}

func chooseConfig[T any](set configset.Set[T]) (T, error) {
	return set.Choose(fxRunOpts.Mode)
}

func Component[T any](name string, set configset.Set[T], opts ...fx.Option) fx.Option {
	config, err := chooseConfig(set)
	if err != nil {
		return fx.Error(fmt.Errorf("%s: %w", name, err))
	}

	return fx.Module(
		name,
		append(
			[]fx.Option{
				fx.Decorate(func(z *zap.Logger) *zap.Logger { return z.Named(name) }),
				fx.Provide(fxGetConfig(name, config), fx.Private),
				fx.Invoke(func(cfg *Config, c *T) { cfg.Store(name, c) }),
			},
			opts...,
		)...,
	)
}

func fxLogger(z *zap.Logger) fxevent.Logger {
	l := &fxevent.ZapLogger{Logger: z.Named("fx")}
	// When Zap writes a log message, it checks that the message's logging level
	// is >= AK's (configurable) logging threshold. "Debug" (-1) is the lowest
	// (i.e. noisiest) possible logging level, so setting FX's info logging to
	// "Debug - 1" means that it will never be >= AK's logging threshold,
	// which means that FX will never log info messages, only errors.
	l.UseLogLevel(zapcore.DebugLevel - 1)
	l.UseErrorLevel(zapcore.ErrorLevel)
	return l
}

func HookOnStart(lc fx.Lifecycle, f func(context.Context) error) { lc.Append(fx.Hook{OnStart: f}) }
func HookOnStop(lc fx.Lifecycle, f func(context.Context) error)  { lc.Append(fx.Hook{OnStop: f}) }

func HookSimpleOnStart(lc fx.Lifecycle, f func()) {
	HookOnStart(lc, func(context.Context) error { f(); return nil })
}

func HookSimpleOnStop(lc fx.Lifecycle, f func()) {
	HookOnStop(lc, func(context.Context) error { f(); return nil })
}
