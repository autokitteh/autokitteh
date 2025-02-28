package svccommon

import (
	"fmt"
	"net/http"
	goruntime "runtime"
	"sync/atomic"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/backend/health/healthchecker"
	"go.autokitteh.dev/autokitteh/internal/backend/health/healthreporter"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/internal/backend/telemetry"
	"go.autokitteh.dev/autokitteh/internal/version"
)

type pprofConfig struct {
	Enable bool `koanf:"enable"`
	Port   int  `koanf:"port"`
}

var pprofConfigs = configset.Set[pprofConfig]{
	Default: &pprofConfig{Enable: true, Port: 6060},
}

// Common components for all services.
func FXCommonComponents(cfg *Config, silent bool) []fx.Option {
	return []fx.Option{
		fx.Supply(cfg),
		LoggerFxOpt(silent),
		Component("telemetry", telemetry.Configs, fx.Provide(telemetry.New)),
		Component(
			"pprof",
			pprofConfigs,
			fx.Invoke(func(cfg *pprofConfig, lc fx.Lifecycle, z *zap.Logger) {
				if !cfg.Enable {
					return
				}

				HookSimpleOnStart(lc, func() {
					go func() {
						addr := fmt.Sprintf("localhost:%d", cfg.Port)
						if err := http.ListenAndServe(addr, nil); err != nil {
							z.Error("listen", zap.Error(err))
						}
					}()
				})
			}),
		),
		Component("healthcheck", configset.Empty, fx.Provide(healthchecker.New)),
		fx.Invoke(func(z *zap.Logger, lc fx.Lifecycle, muxes *muxes.Muxes, h healthreporter.HealthReporter) {
			var ready atomic.Bool

			healthz := func(w http.ResponseWriter, r *http.Request) {
				if err := h.Report(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusOK)
			}

			muxes.NoAuth.HandleFunc("GET /healthz", healthz)
			muxes.NoAuth.HandleFunc("GET /internal/healthz", healthz)

			readyz := func(w http.ResponseWriter, r *http.Request) {
				if !ready.Load() {
					w.WriteHeader(http.StatusServiceUnavailable)
					return
				}

				w.WriteHeader(http.StatusOK)
			}

			muxes.NoAuth.HandleFunc("GET /readyz", readyz)
			muxes.NoAuth.HandleFunc("GET /internal/readyz", readyz)

			HookSimpleOnStart(lc, func() {
				ready.Store(true)
				z.Info("ready", zap.String("version", version.Version), zap.String("id", fixtures.ProcessID()), zap.Int("gomaxprocs", goruntime.GOMAXPROCS(0)))
			})
		}),
	}
}
