package svc

import (
	"context"
	"fmt"
	"os"
	"strings"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

var components = kittehs.FilterZeroes(strings.Split(os.Getenv("AK_COMPONENTS"), ","))

func isComponentEnabled(name string) bool {
	if len(components) == 0 {
		return true
	}

	enabled := false

	for _, c := range components {
		if c == "" {
			continue
		}

		if c[0] == '-' {
			if len(c) == 1 || c[1:] == name || c[1:] == "*" {
				enabled = false
			}

			continue
		}

		if c == name || c == "*" {
			enabled = true
		}
	}

	return enabled
}

// TODO: need this so RunOptions would not needed to be passed every time Component is called. This is ugly, fix.
var fxRunOpts RunOptions

func setFXRunOpts(opts RunOptions) { fxRunOpts = opts }

func fxGetConfig[T any](path string, def T) func(c *Config) (*T, error) {
	return func(c *Config) (*T, error) {
		return GetConfig(c, path, def)
	}
}

func chooseConfig[T any](set configset.Set[T]) (T, error) {
	return set.Choose(configset.Mode(fxRunOpts.Mode))
}

func Invoke(name string, funcs ...any) fx.Option {
	if !isComponentEnabled(name) {
		return fx.Module(name, fx.Invoke(func(sl *zap.SugaredLogger) { sl.Info("disabled") }))
	}

	return fx.Invoke(funcs...)
}

func Component[T any](name string, set configset.Set[T], opts ...fx.Option) fx.Option {
	if !isComponentEnabled(name) {
		return fx.Module(name, fx.Invoke(func(sl *zap.SugaredLogger) { sl.Info("disabled") }))
	}

	config, err := chooseConfig(set)
	if err != nil {
		return fx.Error(fmt.Errorf("%s: %w", name, err))
	}

	return fx.Module(
		name,
		append(
			[]fx.Option{
				fx.Decorate(func(l *zap.Logger) *zap.Logger { return l.Named(name) }),
				fx.Decorate(func(sl *zap.SugaredLogger) *zap.SugaredLogger { return sl.Named(name) }),
				fx.Provide(fxGetConfig(name, config), fx.Private),
				fx.Invoke(func(cfg *Config, c *T) { cfg.Store(name, c) }),
			},
			opts...,
		)...,
	)
}

func fxLogger(z *zap.Logger) fxevent.Logger {
	l := &fxevent.ZapLogger{Logger: z.Named("fx")}
	l.UseLogLevel(zapcore.DebugLevel)
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
