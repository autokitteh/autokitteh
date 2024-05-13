package svc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/fatih/color"
	"go.temporal.io/sdk/client"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/backend/runtimes"
	"go.autokitteh.dev/autokitteh/internal/backend/applygrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authhttpmiddleware"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authjwttokens"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authloginhttpsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authsessions"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/builds"
	"go.autokitteh.dev/autokitteh/internal/backend/buildsgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/connections"
	"go.autokitteh.dev/autokitteh/internal/backend/connectionsgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/dashboardsvc"
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
	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/backend/httpsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/integrations"
	"go.autokitteh.dev/autokitteh/internal/backend/integrationsgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/logger"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/internal/backend/oauth"
	"go.autokitteh.dev/autokitteh/internal/backend/projects"
	"go.autokitteh.dev/autokitteh/internal/backend/projectsgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/secrets"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions"
	"go.autokitteh.dev/autokitteh/internal/backend/sessionsgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/store"
	"go.autokitteh.dev/autokitteh/internal/backend/storegrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/internal/backend/triggers"
	"go.autokitteh.dev/autokitteh/internal/backend/triggersgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/vars"
	"go.autokitteh.dev/autokitteh/internal/backend/varsgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/webtools"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/version"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/auth/v1/authv1connect"
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
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1/sessionsv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/triggers/v1/triggersv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/vars/v1/varsv1connect"
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
	)
}

type HTTPServerAddr string

func makeFxOpts(cfg *Config, opts RunOptions) []fx.Option {
	return []fx.Option{
		fx.Supply(cfg),
		LoggerFxOpt(),
		fx.Invoke(func(lc fx.Lifecycle, db db.DB) { HookOnStart(lc, db.Setup) }),

		Component("auth", configset.Empty, fx.Provide(authsvc.New)),
		Component("authjwttokens", authjwttokens.Configs, fx.Provide(authjwttokens.New)),
		Component("authsessions", authsessions.Configs, fx.Provide(authsessions.New)),
		Component("authhttmiddleware", authhttpmiddleware.Configs, fx.Provide(authhttpmiddleware.New)),

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
			fx.Invoke(func(lc fx.Lifecycle, c temporalclient.Client) {
				HookOnStart(lc, c.Start)
				HookOnStop(lc, c.Stop)
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
		Component("vars", configset.Empty, fx.Provide(vars.New)),
		Component("secrets", secrets.Configs, fx.Provide(secrets.New)),
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
			fx.Invoke(func(lc fx.Lifecycle, muxes *muxes.Muxes, t webtools.Svc) {
				t.Init(muxes)
				HookOnStart(lc, t.Setup)
			}),
		),
		fx.Provide(func(s fxServices) sdkservices.Services { return &s }),
		fx.Invoke(authgrpcsvc.Init),
		fx.Invoke(applygrpcsvc.Init),
		fx.Invoke(buildsgrpcsvc.Init),
		fx.Invoke(connectionsgrpcsvc.Init),
		fx.Invoke(deploymentsgrpcsvc.Init),
		fx.Invoke(dispatchergrpcsvc.Init),
		fx.Invoke(envsgrpcsvc.Init),
		fx.Invoke(eventsgrpcsvc.Init),
		fx.Invoke(integrationsgrpcsvc.Init),
		fx.Invoke(oauth.Init),
		fx.Invoke(projectsgrpcsvc.Init),
		fx.Invoke(func(z *zap.Logger, runtimes sdkservices.Runtimes, muxes *muxes.Muxes) {
			sdkruntimessvc.Init(z, runtimes, muxes.Auth)
		}),
		fx.Invoke(sessionsgrpcsvc.Init),
		fx.Invoke(storegrpcsvc.Init),
		fx.Invoke(triggersgrpcsvc.Init),
		fx.Invoke(varsgrpcsvc.Init),

		Component(
			"http",
			httpsvc.Configs,
			fx.Provide(func(lc fx.Lifecycle, z *zap.Logger, cfg *httpsvc.Config, wrapAuth authhttpmiddleware.AuthMiddlewareDecorator) (svc httpsvc.Svc, all *muxes.Muxes, err error) {
				svc, err = httpsvc.New(
					lc, z, cfg,
					[]string{
						authv1connect.AuthServiceName,
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
						sessionsv1connect.SessionsServiceName,
						triggersv1connect.TriggersServiceName,
						varsv1connect.VarsServiceName,
					},
					[]httpsvc.RequestLogExtractor{
						func(r *http.Request) []zap.Field {
							if user := authcontext.GetAuthnUser(r.Context()); user.IsValid() {
								return []zap.Field{zap.String("user", user.Title())}
							}

							return nil
						},
					},
				)
				if err != nil {
					return
				}

				// Replace the original mux with a main mux and also expose the original mux as the no-auth one.
				authMux := http.NewServeMux()

				mux := svc.Mux()
				mux.Handle("/", wrapAuth(authMux))

				all = &muxes.Muxes{Auth: authMux, NoAuth: mux}

				return
			}),
		),
		Component("authloginhttpsvc", authloginhttpsvc.Configs, fx.Invoke(authloginhttpsvc.Init)),
		fx.Invoke(func(muxes *muxes.Muxes, h integrationsweb.Handler) {
			muxes.NoAuth.Handle("/i/", &h)
		}),
		fx.Invoke(dashboardsvc.Init),
		fx.Invoke(oauth.InitWebhook),
		Component("integrations", integrations.Configs, fx.Provide(integrations.New)),
		fx.Invoke(func(lc fx.Lifecycle, l *zap.Logger, muxes *muxes.Muxes, vs sdkservices.Vars, o sdkservices.OAuth, d dispatcher.Dispatcher, c sdkservices.Connections, p sdkservices.Projects) {
			HookOnStart(lc, func(ctx context.Context) error {
				return integrations.Start(ctx, l, muxes.NoAuth, vs, o, d, c, p)
			})
		}),
		fx.Invoke(func(z *zap.Logger, muxes *muxes.Muxes) {
			srv := http.StripPrefix("/static/", http.FileServer(http.FS(static.RootWebContent)))
			muxes.NoAuth.Handle("/static/", srv)
			muxes.NoAuth.Handle("/favicon-32x32.png", srv)
			muxes.NoAuth.Handle("/favicon-16x16.png", srv)
			muxes.NoAuth.Handle("/favicon.ico", srv)
		}),
		Component(
			"banner",
			bannerConfigs,
			fx.Invoke(func(cfg *bannerConfig, lc fx.Lifecycle, z *zap.Logger, httpsvc httpsvc.Svc, tclient temporalclient.Client) {
				HookSimpleOnStart(lc, func() {
					temporalFrontendAddr, temporalUIAddr := tclient.TemporalAddr()
					printBanner(cfg, opts, httpsvc.Addr(), temporalFrontendAddr, temporalUIAddr)
				})
			}),
		),
		fx.Invoke(func(muxes *muxes.Muxes) {
			t0 := time.Now()

			muxes.NoAuth.HandleFunc("/id", func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, fixtures.ProcessID())
			})
			muxes.NoAuth.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				kittehs.Must0(json.NewEncoder(w).Encode(version.Version))
			})
			muxes.NoAuth.HandleFunc("/uptime", func(w http.ResponseWriter, r *http.Request) {
				uptime := time.Since(t0)

				resp := struct {
					Text    string `json:"text"`
					Seconds uint64 `json:"seconds"`
				}{
					Text:    uptime.String(),
					Seconds: uint64(uptime.Seconds()),
				}

				w.Header().Set("Content-Type", "application/json")
				kittehs.Must0(json.NewEncoder(w).Encode(resp))
			})
		}),
		fx.Invoke(func(z *zap.Logger, lc fx.Lifecycle, muxes *muxes.Muxes) {
			var ready atomic.Bool

			muxes.NoAuth.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
				// TODO(ENG-530): check db, temporal, etc.
				w.WriteHeader(http.StatusOK)
			})

			muxes.NoAuth.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
				if !ready.Load() {
					w.WriteHeader(http.StatusServiceUnavailable)
					return
				}

				w.WriteHeader(http.StatusOK)
			})

			HookSimpleOnStart(lc, func() {
				ready.Store(true)
				z.Info("ready", zap.String("version", version.Version), zap.String("id", fixtures.ProcessID()))
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

	if !ropts.Mode.IsDefault() && !ropts.Silent {
		printModeWarning(ropts.Mode)
	}

	var svcs sdkservices.Services = &fxServices{}

	return append(opts, fx.Populate(svcs))
}

func StartDB(ctx context.Context, cfg *Config, ropt RunOptions) (db.DB, error) {
	setFXRunOpts(ropt)

	var db db.DB

	if err := fx.New(
		fx.Supply(cfg),
		LoggerFxOpt(),
		DBFxOpt(),
		fx.Populate(&db),
	).Start(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
