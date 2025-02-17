package svc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	goruntime "runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/fatih/color"
	"go.temporal.io/sdk/client"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/applygrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authhttpmiddleware"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authloginhttpsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authsessions"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authtokens/authjwttokens"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authz"
	"go.autokitteh.dev/autokitteh/internal/backend/builds"
	"go.autokitteh.dev/autokitteh/internal/backend/buildsgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/connections"
	"go.autokitteh.dev/autokitteh/internal/backend/connectionsgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/connectionsinitsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/cron"
	"go.autokitteh.dev/autokitteh/internal/backend/dashboardsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbfactory"
	"go.autokitteh.dev/autokitteh/internal/backend/deployments"
	"go.autokitteh.dev/autokitteh/internal/backend/deploymentsgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/dispatcher"
	"go.autokitteh.dev/autokitteh/internal/backend/dispatchergrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/events"
	"go.autokitteh.dev/autokitteh/internal/backend/eventsgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/backend/health/healthchecker"
	"go.autokitteh.dev/autokitteh/internal/backend/health/healthreporter"
	"go.autokitteh.dev/autokitteh/internal/backend/httpsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/integrationsgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/logger"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/internal/backend/oauth"
	"go.autokitteh.dev/autokitteh/internal/backend/orgs"
	"go.autokitteh.dev/autokitteh/internal/backend/orgsgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/policy/opapolicy"
	"go.autokitteh.dev/autokitteh/internal/backend/projects"
	"go.autokitteh.dev/autokitteh/internal/backend/projectsgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/scheduler"
	"go.autokitteh.dev/autokitteh/internal/backend/secrets"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions"
	"go.autokitteh.dev/autokitteh/internal/backend/sessionsgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/store"
	"go.autokitteh.dev/autokitteh/internal/backend/storegrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/telemetry"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/internal/backend/triggers"
	"go.autokitteh.dev/autokitteh/internal/backend/triggersgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/users"
	"go.autokitteh.dev/autokitteh/internal/backend/usersgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/vars"
	"go.autokitteh.dev/autokitteh/internal/backend/varsgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/webhookssvc"
	"go.autokitteh.dev/autokitteh/internal/backend/webplatform"
	"go.autokitteh.dev/autokitteh/internal/backend/webtools"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/version"
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

func LoggerFxOpt(silent bool) fx.Option {
	if silent {
		return fx.Supply(zap.NewNop())
	}

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
			fx.Invoke(func(lc fx.Lifecycle, l *zap.Logger, gdb db.DB, cfg *svcConfig) error {
				HookOnStart(lc, gdb.Connect)
				HookOnStart(lc, gdb.Setup)

				if path := cfg.SeedObjectsPath; path != "" {
					bs, err := os.ReadFile(path)
					if err != nil {
						return fmt.Errorf("read seed objects: %w", err)
					}

					var objs []sdktypes.AnyObject

					if err := json.Unmarshal(bs, &objs); err != nil {
						return fmt.Errorf("unmarshal seed objects: %w", err)
					}

					HookOnStart(lc, func(ctx context.Context) error {
						if len(objs) == 0 {
							return nil
						}

						l.Info("populating db with seed objects", zap.Int("n", len(objs)), zap.String("path", path))
						return db.Populate(ctx, gdb, kittehs.Transform(objs, sdktypes.UnwrapAnyObject)...)
					})
				}

				return nil
			}),
		),
	)
}

type pprofConfig struct {
	Enable bool `koanf:"enable"`
	Port   int  `koanf:"port"`
}

var pprofConfigs = configset.Set[pprofConfig]{
	Default: &pprofConfig{Enable: true, Port: 6060},
}

type HTTPServerAddr string

func makeFxOpts(cfg *Config, opts RunOptions) []fx.Option {
	return []fx.Option{
		fx.Supply(cfg),

		SupplyConfig("svc", svcConfigs),

		LoggerFxOpt(opts.Silent),
		DBFxOpt(),

		Component("auth", configset.Empty, fx.Provide(authsvc.New)),
		Component("authjwttokens", authjwttokens.Configs, fx.Provide(authjwttokens.New)),
		Component("authsessions", authsessions.Configs, fx.Provide(authsessions.New)),
		Component(
			"authhttpmiddleware",
			configset.Empty,
			fx.Provide(authhttpmiddleware.New),
			fx.Provide(authhttpmiddleware.AuthorizationHeaderExtractor),
		),
		Component("orgs", configset.Empty, fx.Provide(orgs.New)),
		Component(
			"users",
			users.Configs,
			fx.Provide(users.New),
			fx.Provide(func(u users.Users) sdkservices.Users { return u }),
			fx.Invoke(func(lc fx.Lifecycle, u users.Users) {
				HookOnStart(lc, u.Setup)
			}),
		),
		Component("opapolicy", opapolicy.Configs, fx.Provide(opapolicy.New)),
		Component("authz", configset.Empty, fx.Provide(authz.NewPolicyCheckFunc)),

		Component(
			"temporalclient",
			temporalclient.Configs,
			fx.Provide(func(cfg *temporalclient.Config, z *zap.Logger) (temporalclient.Client, error) {
				if opts.TemporalClient == nil {
					return temporalclient.New(cfg, z)
				}

				return temporalclient.NewFromTemporalClient(&cfg.Monitor, z, opts.TemporalClient)
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
		Component("deployments", deployments.Configs, fx.Provide(deployments.New)),
		Component("projects", configset.Empty, fx.Provide(projects.New)),
		Component("projectsgrpcsvc", projectsgrpcsvc.Configs, fx.Provide(projectsgrpcsvc.New)),
		Component(
			"vars",
			vars.Configs,
			fx.Provide(vars.New),
			fx.Provide(func(v *vars.Vars) sdkservices.Vars { return v }),
			fx.Invoke(func(v *vars.Vars, c sdkservices.Connections) { v.SetConnections(c) }),
		),
		Component("secrets", secrets.Configs, fx.Provide(secrets.New)),
		Component("events", configset.Empty, fx.Provide(events.New)),
		Component("triggers", configset.Empty, fx.Provide(triggers.New)),
		Component("oauth", configset.Empty, fx.Provide(oauth.New)),
		runtimesFXOption(),
		Component("healthcheck", configset.Empty, fx.Provide(healthchecker.New)),
		Component(
			"scheduler",
			scheduler.Configs,
			fx.Provide(scheduler.New),
			fx.Invoke(
				func(lc fx.Lifecycle, sch *scheduler.Scheduler, d sdkservices.Dispatcher, ts sdkservices.Triggers) {
					HookOnStart(lc, func(ctx context.Context) error {
						return sch.Start(ctx, d, ts)
					})
				},
			),
		),
		Component(
			"cron",
			cron.Configs,
			fx.Provide(cron.New),
			fx.Invoke(
				func(lc fx.Lifecycle, ct *cron.Cron, c sdkservices.Connections, v sdkservices.Vars, o sdkservices.OAuth) {
					HookOnStart(lc, func(ctx context.Context) error {
						return ct.Start(ctx, c, v, o)
					})
				},
			),
		),
		Component(
			"dispatcher",
			dispatcher.Configs,
			fx.Provide(func(lc fx.Lifecycle, l *zap.Logger, cfg *dispatcher.Config, svcs dispatcher.Svcs) (sdkservices.Dispatcher, sdkservices.DispatchFunc) {
				d := dispatcher.New(l, cfg, svcs)
				HookOnStart(lc, d.Start)
				return d, d.Dispatch
			}),
		),
		Component(
			"webtools",
			webtools.Configs,
			fx.Provide(webtools.New),
			fx.Invoke(func(lc fx.Lifecycle, muxes *muxes.Muxes, svc webtools.Svc) {
				svc.Init(muxes)
				HookOnStart(lc, svc.Setup)
			}),
		),
		Component(
			"webplatform",
			webplatform.Configs,
			fx.Provide(webplatform.New),
			fx.Invoke(func(lc fx.Lifecycle, svc *webplatform.Svc) {
				HookOnStart(lc, svc.Start)
				HookOnStop(lc, svc.Stop)
			}),
		),
		fx.Provide(func(s sdkservices.ServicesStruct) sdkservices.Services { return &s }),
		fx.Invoke(authgrpcsvc.Init),
		fx.Invoke(applygrpcsvc.Init),
		fx.Invoke(buildsgrpcsvc.Init),
		fx.Invoke(connectionsgrpcsvc.Init),
		fx.Invoke(deploymentsgrpcsvc.Init),
		fx.Invoke(dispatchergrpcsvc.Init),
		fx.Invoke(eventsgrpcsvc.Init),
		fx.Invoke(usersgrpcsvc.Init),
		fx.Invoke(orgsgrpcsvc.Init),
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
		Component("telemetry", telemetry.Configs, fx.Provide(telemetry.New)),
		Component(
			"http",
			httpsvc.Configs,
			fx.Provide(func(lc fx.Lifecycle, z *zap.Logger, cfg *httpsvc.Config,
				authzCheckFunc authz.CheckFunc,
				wrapAuth authhttpmiddleware.AuthMiddlewareDecorator,
				authHdrExtractor authhttpmiddleware.AuthHeaderExtractor, telemetry *telemetry.Telemetry,
			) (svc httpsvc.Svc, all *muxes.Muxes, err error) {
				svc, err = httpsvc.New(
					lc, z, cfg,
					authzCheckFunc,
					[]httpsvc.RequestLogExtractor{
						// Note: auth middleware will be connected after interceptor, so in httpsvc and interceptor handler
						// there is (still) no parsed user in the httpRequest context. So in order to log the user in the
						// same place where httpRequest is logged we need to extract it from the header
						func(r *http.Request) []zap.Field {
							if uid := authHdrExtractor(r); uid.IsValid() {
								return []zap.Field{zap.String("user_id", uid.String())}
							}

							return nil
						},
					},
					telemetry,
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
			muxes.NoAuth.Handle("GET /i/{$}", &h)
		}),
		fx.Invoke(dashboardsvc.Init),
		fx.Invoke(oauth.InitWebhook),
		fx.Invoke(connectionsinitsvc.Init),
		Component(
			"webhooks",
			configset.Empty,
			fx.Provide(webhookssvc.New),
			fx.Invoke(func(lc fx.Lifecycle, w *webhookssvc.Service, muxes *muxes.Muxes) {
				HookSimpleOnStart(lc, func() { w.Start(muxes) })
			}),
		),
		integrationsFXOption(),
		fx.Invoke(func(z *zap.Logger, muxes *muxes.Muxes) {
			srv := http.FileServer(http.FS(static.RootWebContent))
			muxes.NoAuth.Handle("GET /static/", http.StripPrefix("/static/", srv))
			muxes.NoAuth.Handle("GET /favicon-16x16.png", srv)
			muxes.NoAuth.Handle("GET /favicon-32x32.png", srv)
			muxes.NoAuth.Handle("GET /favicon.ico", srv)
			muxes.NoAuth.Handle("GET /robots.txt", srv)
			muxes.NoAuth.Handle("GET /site.webmanifest", srv)
		}),
		Component(
			"banner",
			bannerConfigs,
			fx.Invoke(func(cfg *bannerConfig, lc fx.Lifecycle, z *zap.Logger, httpsvc httpsvc.Svc, tclient temporalclient.Client, wp *webplatform.Svc) {
				HookSimpleOnStart(lc, func() {
					temporalFrontendAddr, temporalUIAddr := tclient.TemporalAddr()
					printBanner(cfg, opts, httpsvc.Addr(), wp.Addr(), wp.Version(), temporalFrontendAddr, temporalUIAddr)
				})
			}),
		),
		fx.Invoke(func(muxes *muxes.Muxes, svcConfig *svcConfig) {
			muxes.NoAuth.Handle("/{$}", http.RedirectHandler(svcConfig.RootRedirect, http.StatusSeeOther))
			muxes.NoAuth.HandleFunc("/internal/{$}", func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `<html><body>
	<ul>
		<li><a href="/internal/id">id</a></li>
		<li><a href="/internal/version">version</a></li>
		<li><a href="/internal/uptime">uptime</a></li>
		<li><a href="/internal/healthz">healthz</a></li>
		<li><a href="/internal/readyz">readyz</a></li>
		<li><a href="/internal/dashboard">dashboard</a></li>
	</ul>
</body></html>`)
			})

			muxes.NoAuth.HandleFunc("GET /internal/id", func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, fixtures.ProcessID())
			})
			muxes.NoAuth.HandleFunc("GET /internal/version", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				kittehs.Must0(json.NewEncoder(w).Encode(version.Version))
			})
			muxes.NoAuth.HandleFunc("GET /internal/uptime", func(w http.ResponseWriter, r *http.Request) {
				uptime := fixtures.Uptime().Truncate(time.Second)

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

	var svcs sdkservices.Services = &sdkservices.ServicesStruct{}

	return append(opts, fx.Populate(svcs))
}

func StartDB(ctx context.Context, cfg *Config, ropt RunOptions) (db.DB, error) {
	setFXRunOpts(ropt)

	var db db.DB

	if err := fx.New(
		fx.Supply(cfg),
		LoggerFxOpt(ropt.Silent),
		DBFxOpt(),
		fx.Populate(&db),
	).Start(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
