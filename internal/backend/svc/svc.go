package svc

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/fatih/color"
	"go.temporal.io/sdk/client"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/backend/runtimes"
	"go.autokitteh.dev/autokitteh/internal/backend/applygrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/builds"
	"go.autokitteh.dev/autokitteh/internal/backend/buildsgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/connections"
	"go.autokitteh.dev/autokitteh/internal/backend/connectionsgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbfactory"
	"go.autokitteh.dev/autokitteh/internal/backend/deployments"
	"go.autokitteh.dev/autokitteh/internal/backend/deploymentsgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/dispatcher"
	"go.autokitteh.dev/autokitteh/internal/backend/dispatchergrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/envs"
	"go.autokitteh.dev/autokitteh/internal/backend/envsgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/events"
	"go.autokitteh.dev/autokitteh/internal/backend/eventsgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/httpsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/integrations"
	"go.autokitteh.dev/autokitteh/internal/backend/integrationsgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/logger"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/internal/backend/oauth"
	"go.autokitteh.dev/autokitteh/internal/backend/projects"
	"go.autokitteh.dev/autokitteh/internal/backend/projectsgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/secrets"
	"go.autokitteh.dev/autokitteh/internal/backend/secretsgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions"
	"go.autokitteh.dev/autokitteh/internal/backend/sessionsgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/store"
	"go.autokitteh.dev/autokitteh/internal/backend/storegrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/internal/backend/triggers"
	"go.autokitteh.dev/autokitteh/internal/backend/triggersgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/webtools"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/builds/v1/buildsv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/connections/v1/connectionsv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/deployments/v1/deploymentsv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/dispatcher/v1/dispatcherv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/envs/v1/envsv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/events/v1/eventsv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/integrations/v1/integrationsv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/oauth/v1/oauthv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/projects/v1/projectsv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runtimes/v1/runtimesv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/secrets/v1/secretsv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1/sessionsv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/triggers/v1/triggersv1connect"
	"go.autokitteh.dev/autokitteh/sdk/sdkruntimessvc"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	integrationsweb "go.autokitteh.dev/autokitteh/web/integrations"
	"go.autokitteh.dev/autokitteh/web/static"
)

var warningColor = color.New(color.FgRed).Add(color.Bold).SprintFunc()

func printModeWarning(mode configset.Mode) {
	fmt.Fprint(
		os.Stderr,
		warningColor(fmt.Sprintf("\n*** %s MODE *** NEVER USE IN PRODUCTION\n\n", strings.ToUpper(string(mode)))),
	)
}

func LoggerFxOpt() fx.Option {
	return fx.Module(
		"logger",
		fx.Provide(fxGetConfig("logger", kittehs.Must1(chooseConfig(logger.Configs)))),
		fx.Provide(logger.New),
	)
}

func DBFxOpt() fx.Option {
	return fx.Options(
		Component(
			"db",
			dbfactory.Configs,
			fx.Provide(dbfactory.New),
			fx.Invoke(func(lc fx.Lifecycle, z *zap.Logger, db db.DB) {
				HookOnStart(lc, db.Connect)
			}),
		),
		fx.Invoke(func(lc fx.Lifecycle, db db.DB) { HookOnStart(lc, db.Setup) }),
	)
}

type HTTPServerAddr string

func makeFxOpts(cfg *Config, opts RunOptions) []fx.Option {
	return []fx.Option{
		fx.Supply(cfg),
		LoggerFxOpt(),
		DBFxOpt(),
		Component(
			"temporalclient",
			temporalclient.Configs,
			fx.Provide(func(cfg *temporalclient.Config, z *zap.Logger) (temporalclient.Client, error) {
				if opts.TemporalClient != nil {
					return temporalclient.NewFromClient(&cfg.Monitor, z, opts.TemporalClient)
				}

				return temporalclient.New(cfg, z)
			}),
			fx.Provide(func(c temporalclient.Client) client.Client { return c.Temporal() }),
			fx.Invoke(func(lc fx.Lifecycle, c temporalclient.Client) {
				HookOnStart(lc, c.Start)
				HookOnStop(lc, c.Stop)
			}),
		),
		Component("secrets", secrets.Configs, fx.Provide(secrets.New)),
		Component(
			"sessions",
			sessions.Configs,
			fx.Provide(sessions.New),
			fx.Provide(func(s sessions.Sessions) sdkservices.Sessions { return s }),
			fx.Invoke(func(lc fx.Lifecycle, s sessions.Sessions) { HookOnStart(lc, s.StartWorkers) }),
		),
		Component("store", store.Configs, fx.Provide(store.New)),
		Component("builds", configset.Empty, fx.Provide(builds.New)),
		Component("connections", configset.Empty, fx.Provide(connections.New)),
		Component("deployments", configset.Empty, fx.Provide(deployments.New)),
		Component("projects", configset.Empty, fx.Provide(projects.New)),
		Component("projectsgrpcsvc", projectsgrpcsvc.Configs, fx.Provide(projectsgrpcsvc.New)),
		Component("envs", configset.Empty, fx.Provide(envs.New)),
		Component("events", configset.Empty, fx.Provide(events.New)),
		Component("triggers", configset.Empty, fx.Provide(triggers.New)),
		Component("oauth", configset.Empty, fx.Provide(oauth.New)),
		Component("runtimes", configset.Empty, fx.Provide(runtimes.New)),
		Component(
			"dispatcher",
			configset.Empty,
			fx.Provide(dispatcher.New),
			fx.Provide(func(d dispatcher.Dispatcher) sdkservices.Dispatcher { return d }),
			fx.Invoke(func(lc fx.Lifecycle, d dispatcher.Dispatcher) { HookOnStart(lc, d.Start) }),
		),
		Component(
			"webtools",
			webtools.Configs,
			fx.Provide(webtools.New),
			fx.Invoke(func(lc fx.Lifecycle, mux *http.ServeMux, t webtools.Svc) {
				t.Init(mux)
				HookOnStart(lc, t.Setup)
			}),
		),
		fx.Provide(func(s fxServices) sdkservices.Services { return &s }),
		fx.Invoke(sdkruntimessvc.Init),
		fx.Invoke(applygrpcsvc.Init),
		fx.Invoke(func(p *projectsgrpcsvc.Server, mux *http.ServeMux) { p.Init(mux) }),
		fx.Invoke(buildsgrpcsvc.Init),
		fx.Invoke(connectionsgrpcsvc.Init),
		fx.Invoke(deploymentsgrpcsvc.Init),
		fx.Invoke(dispatchergrpcsvc.Init),
		fx.Invoke(envsgrpcsvc.Init),
		fx.Invoke(eventsgrpcsvc.Init),
		fx.Invoke(integrationsgrpcsvc.Init),
		fx.Invoke(triggersgrpcsvc.Init),
		fx.Invoke(oauth.Init),
		fx.Invoke(secretsgrpcsvc.Init),
		fx.Invoke(storegrpcsvc.Init),
		fx.Invoke(sessionsgrpcsvc.Init),
		Component(
			"http",
			httpsvc.Configs,
			fx.Provide(func(lc fx.Lifecycle, z *zap.Logger, cfg *httpsvc.Config) (svc httpsvc.Svc, mux *http.ServeMux, all *muxes.Muxes, err error) {
				svc, err = httpsvc.New(
					lc, z, cfg,
					[]string{
						buildsv1connect.BuildsServiceName,
						connectionsv1connect.ConnectionsServiceName,
						deploymentsv1connect.DeploymentsServiceName,
						dispatcherv1connect.DispatcherServiceName,
						envsv1connect.EnvsServiceName,
						eventsv1connect.EventsServiceName,
						integrationsv1connect.IntegrationsServiceName,
						oauthv1connect.OAuthServiceName,
						projectsv1connect.ProjectsServiceName,
						runtimesv1connect.RuntimesServiceName,
						secretsv1connect.SecretsServiceName,
						sessionsv1connect.SessionsServiceName,
						triggersv1connect.TriggersServiceName,
					},
					nil,
				)
				if err != nil {
					return
				}

				// Replace the original mux with a main mux and also expose the original mux as the no-auth one.
				authed := http.NewServeMux()

				wrap := func(h http.Handler) http.Handler {
					// TODO(ENG-3): Wrap the handler with auth middleware.
					return h
				}

				mux = svc.Mux()

				mux.Handle("/", wrap(authed))

				all = &muxes.Muxes{Auth: authed, NoAuth: mux}

				return
			}),
		),
		fx.Invoke(func(mux *http.ServeMux, h integrationsweb.Handler) {
			mux.Handle("/i/", &h)
		}),
		fx.Invoke(func(mux *http.ServeMux, l *zap.Logger, s sdkservices.Services) {
			mux.Handle("/oauth/", oauth.NewWebhook(l, s))
		}),
		Component("integrations", integrations.Configs, fx.Provide(integrations.New)),
		fx.Invoke(func(lc fx.Lifecycle, l *zap.Logger, muxes *muxes.Muxes, s sdkservices.Secrets, o sdkservices.OAuth, d dispatcher.Dispatcher) {
			HookOnStart(lc, func(ctx context.Context) error {
				return integrations.Start(ctx, l, muxes.NoAuth, s, o, d)
			})
		}),
		indexOption(),
		fx.Invoke(func(z *zap.Logger, mux *http.ServeMux) {
			srv := http.StripPrefix("/static/", http.FileServer(http.FS(static.RootWebContent)))
			mux.Handle("/static/", srv)
			mux.Handle("/favicon-32x32.png", srv)
			mux.Handle("/favicon-16x16.png", srv)
		}),
		fx.Invoke(func(lc fx.Lifecycle, z *zap.Logger, httpsvc httpsvc.Svc, tclient temporalclient.Client) {
			HookSimpleOnStart(lc, func() {
				temporalFrontendAddr, temporalUIAddr := tclient.TemporalAddr()
				printBanner(opts, httpsvc.Addr(), temporalFrontendAddr, temporalUIAddr)
			})
		}),
		fx.Invoke(func(z *zap.Logger, lc fx.Lifecycle, mux *http.ServeMux) {
			var ready atomic.Bool

			mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
				// TODO(ENG-530): check db, temporal, etc.
				w.WriteHeader(http.StatusOK)
			})

			mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
				if !ready.Load() {
					w.WriteHeader(http.StatusServiceUnavailable)
					return
				}

				w.WriteHeader(http.StatusOK)
			})

			HookSimpleOnStart(lc, func() {
				ready.Store(true)
				z.Info("ready")
			})
		}),
	}
}

type RunOptions struct {
	Mode           configset.Mode
	Silent         bool          // No logs at all
	TemporalClient client.Client // use this instead of creating a new temporal client.
}

func NewOpts(cfg *Config, ropts RunOptions) []fx.Option {
	setFXRunOpts(ropts)

	opts := makeFxOpts(cfg, ropts)

	if ropts.Silent {
		opts = append(opts, fx.NopLogger)
	} else {
		opts = append(opts, fx.WithLogger(fxLogger))
	}

	if ropts.Mode.IsTest() {
		sdktypes.SetIDGenerator(sdktypes.NewSequentialIDGeneratorForTesting(0))
	}

	if !ropts.Mode.IsDefault() {
		printModeWarning(ropts.Mode)
	}

	var svcs sdkservices.Services = &fxServices{}

	return append(opts, fx.Populate(svcs))
}

func StartDB(ctx context.Context, cfg *Config, ropts RunOptions) (db.DB, error) {
	setFXRunOpts(ropts)

	var db db.DB

	if err := fx.New(
		fx.NopLogger,
		fx.Supply(cfg),
		LoggerFxOpt(),
		DBFxOpt(),
		fx.Populate(&db),
	).Start(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
