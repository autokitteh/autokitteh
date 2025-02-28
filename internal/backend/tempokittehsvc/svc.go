package tempokittehsvc

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/httpsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessioncalls"
	"go.autokitteh.dev/autokitteh/internal/backend/svccommon"
	"go.autokitteh.dev/autokitteh/internal/backend/telemetry"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/tempokitteh"
)

var warningColor = color.New(color.FgRed).Add(color.Bold).SprintFunc()

func printModeWarning(mode configset.Mode) {
	fmt.Fprint(
		os.Stderr,
		warningColor(fmt.Sprintf("\n*** %s MODE *** NEVER USE IN PRODUCTION\n\n", strings.ToUpper(string(mode)))),
	)
}

func makeFxOpts(cfg *svccommon.Config, opts RunOptions) []fx.Option {
	return append(
		svccommon.FXCommonComponents(cfg, opts.Silent),
		svccommon.Component(
			"http",
			httpsvc.Configs,
			fx.Provide(func(lc fx.Lifecycle, z *zap.Logger, cfg *httpsvc.Config,
				telemetry *telemetry.Telemetry,
			) (svc httpsvc.Svc, all *muxes.Muxes, err error) {
				svc, err = httpsvc.New(lc, z, cfg, nil, nil, telemetry)
				if err != nil {
					return
				}

				mux := svc.Mux()

				all = &muxes.Muxes{Auth: mux, NoAuth: mux}

				return
			}),
		),
		svccommon.FXRuntimes(),
		svccommon.Component(
			"temporalclient",
			temporalclient.Configs,
			fx.Provide(func(cfg *temporalclient.Config, l *zap.Logger) (temporalclient.Client, error) {
				c, err := temporalclient.New(cfg, l)
				if err != nil {
					return nil, err
				}

				if err := c.Start(context.Background()); err != nil {
					return nil, err
				}

				return c, nil
			}),
			fx.Invoke(func(lc fx.Lifecycle, c temporalclient.Client) {
				svccommon.HookOnStop(lc, c.Stop)
			}),
			fx.Invoke(func(cfg *temporalclient.Config, tclient temporalclient.Client, muxes *muxes.Muxes) {
				if !cfg.EnableHelperRedirect {
					return
				}

				muxes.NoAuth.HandleFunc("/temporal", func(w http.ResponseWriter, r *http.Request) {
					_, uiAddr := tclient.TemporalAddr()
					if uiAddr == "" {
						return
					}

					http.Redirect(w, r, uiAddr, http.StatusFound)
				})
			}),
		),
		svccommon.Component(
			"sessioncalls",
			configset.Empty,
			fx.Provide(func(l *zap.Logger) sessioncalls.Calls {
				return sessioncalls.New(l, sessioncalls.Config{}, nil)
			}),
			fx.Invoke(func(lc fx.Lifecycle, s sessioncalls.Calls, tc temporalclient.Client) {
				svccommon.HookOnStart(lc, func(ctx context.Context) error {
					return s.StartWorkers(ctx, tc.TemporalClient())
				})
			}),
		),
		svccommon.Component(
			"worker",
			configset.Empty,
			fx.Provide(func(l *zap.Logger, tc temporalclient.Client) tempokitteh.Worker {
				return tempokitteh.NewWorker(
					l.Named("worker"),
					tc.TemporalClient(),
					tempokitteh.WorkerConfig{
						TaskQueueName: opts.TKQueueName,
						WorkerConfig:  temporalclient.WorkerConfig{WorkflowDeadlockTimeout: 10 * time.Second},
					},
				)
			}),
			fx.Invoke(func(lc fx.Lifecycle, w tempokitteh.Worker) {
				svccommon.HookOnStart(lc, func(ctx context.Context) error { return w.Start() })
			}),
		),
	)
}

type RunOptions struct {
	Mode   configset.Mode
	Silent bool // No logs at all

	TKQueueName string
}

func NewOpts(cfg *svccommon.Config, ropts RunOptions) []fx.Option {
	svccommon.SetMode(ropts.Mode)

	opts := makeFxOpts(cfg, ropts)

	if ropts.Silent {
		opts = append(opts, fx.NopLogger)
	} else {
		opts = append(opts, fx.WithLogger(svccommon.FXLogger))
	}

	if ropts.Mode.IsTest() {
		sdktypes.SetIDGenerator(sdktypes.NewSequentialIDGeneratorForTesting(0))
	}

	if !ropts.Mode.IsDefault() && !ropts.Silent {
		printModeWarning(ropts.Mode)
	}

	return opts
}
