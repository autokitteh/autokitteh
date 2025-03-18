package aksvc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
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
	"go.autokitteh.dev/autokitteh/internal/backend/httpsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/integrationsgrpcsvc"
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
	"go.autokitteh.dev/autokitteh/internal/backend/svccommon"
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

func DBFxOpt() fx.Option {
	return fx.Options(
		svccommon.Component(
			"db",
			dbfactory.Configs,
			fx.Provide(dbfactory.New),
			fx.Invoke(func(lc fx.Lifecycle, l *zap.Logger, gdb db.DB, cfg *svcConfig) error {
				svccommon.HookOnStart(lc, gdb.Connect)
				svccommon.HookOnStart(lc, gdb.Setup)

				if path := cfg.SeedObjectsPath; path != "" {
					bs, err := os.ReadFile(path)
					if err != nil {
						return fmt.Errorf("read seed objects: %w", err)
					}

					var objs []sdktypes.AnyObject

					if err := json.Unmarshal(bs, &objs); err != nil {
						return fmt.Errorf("unmarshal seed objects: %w", err)
					}

					svccommon.HookOnStart(lc, func(ctx context.Context) error {
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

type HTTPServerAddr string

func makeFxOpts(cfg *svccommon.Config, opts RunOptions) []fx.Option {
	return append(svccommon.FXCommonComponents(cfg, false),
		svccommon.SupplyConfig("svc", svcConfigs),

		DBFxOpt(),

		svccommon.Component("auth", configset.Empty, fx.Provide(authsvc.New)),
		svccommon.Component("authjwttokens", authjwttokens.Configs, fx.Provide(authjwttokens.New)),
		svccommon.Component("authsessions", authsessions.Configs, fx.Provide(authsessions.New)),
		svccommon.Component(
			"authhttpmiddleware",
			configset.Empty,
			fx.Provide(authhttpmiddleware.New),
			fx.Provide(authhttpmiddleware.AuthorizationHeaderExtractor),
		),
		svccommon.Component("orgs", configset.Empty, fx.Provide(orgs.New)),
		svccommon.Component(
			"users",
			users.Configs,
			fx.Provide(users.New),
			fx.Provide(func(u users.Users) sdkservices.Users { return u }),
			fx.Invoke(func(lc fx.Lifecycle, u users.Users) {
				svccommon.HookOnStart(lc, u.Setup)
			}),
		),
		svccommon.Component("opapolicy", opapolicy.Configs, fx.Provide(opapolicy.New)),
		svccommon.Component("authz", configset.Empty, fx.Provide(authz.NewPolicyCheckFunc)),

		svccommon.Component(
			"temporalclient",
			temporalclient.Configs,
			fx.Provide(func(cfg *temporalclient.Config, z *zap.Logger) (temporalclient.Client, error) {
				if opts.TemporalClient == nil {
					return temporalclient.New(cfg, z)
				}

				return temporalclient.NewFromTemporalClient(&cfg.Monitor, z, opts.TemporalClient)
			}),
			fx.Invoke(func(lc fx.Lifecycle, c temporalclient.Client) {
				svccommon.HookOnStart(lc, c.Start)
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
			"sessions",
			sessions.Configs,
			fx.Provide(sessions.New),
			fx.Provide(func(s sessions.Sessions) sdkservices.Sessions { return s }),
			fx.Invoke(func(lc fx.Lifecycle, s sessions.Sessions) { svccommon.HookOnStart(lc, s.StartWorkers) }),
		),
		svccommon.Component("builds", configset.Empty, fx.Provide(builds.New)),
		svccommon.Component("connections", configset.Empty, fx.Provide(connections.New)),
		svccommon.Component("deployments", deployments.Configs, fx.Provide(deployments.New)),
		svccommon.Component("projects", configset.Empty, fx.Provide(projects.New)),
		svccommon.Component("projectsgrpcsvc", projectsgrpcsvc.Configs, fx.Provide(projectsgrpcsvc.New)),
		svccommon.Component(
			"vars",
			vars.Configs,
			fx.Provide(vars.New),
			fx.Provide(func(v *vars.Vars) sdkservices.Vars { return v }),
			fx.Invoke(func(v *vars.Vars, c sdkservices.Connections) { v.SetConnections(c) }),
		),
		svccommon.Component("secrets", secrets.Configs, fx.Provide(secrets.New)),
		svccommon.Component("events", configset.Empty, fx.Provide(events.New)),
		svccommon.Component("triggers", configset.Empty, fx.Provide(triggers.New)),
		svccommon.Component("oauth", configset.Empty, fx.Provide(oauth.New)),
		svccommon.FXRuntimes(),
		svccommon.Component(
			"scheduler",
			scheduler.Configs,
			fx.Provide(scheduler.New),
			fx.Invoke(
				func(lc fx.Lifecycle, sch *scheduler.Scheduler, d sdkservices.Dispatcher, ts sdkservices.Triggers) {
					svccommon.HookOnStart(lc, func(ctx context.Context) error {
						return sch.Start(ctx, d, ts)
					})
				},
			),
		),
		svccommon.Component(
			"cron",
			cron.Configs,
			fx.Provide(cron.New),
			fx.Invoke(
				func(lc fx.Lifecycle, ct *cron.Cron, c sdkservices.Connections, v sdkservices.Vars, o sdkservices.OAuth) {
					svccommon.HookOnStart(lc, func(ctx context.Context) error {
						return ct.Start(ctx, c, v, o)
					})
				},
			),
		),
		svccommon.Component(
			"dispatcher",
			dispatcher.Configs,
			fx.Provide(func(lc fx.Lifecycle, l *zap.Logger, cfg *dispatcher.Config, svcs dispatcher.Svcs) (sdkservices.Dispatcher, sdkservices.DispatchFunc) {
				d := dispatcher.New(l, cfg, svcs)
				svccommon.HookOnStart(lc, d.Start)
				return d, d.Dispatch
			}),
		),
		svccommon.Component(
			"webtools",
			webtools.Configs,
			fx.Provide(webtools.New),
			fx.Invoke(func(lc fx.Lifecycle, muxes *muxes.Muxes, web webtools.Svc) {
				web.Init(muxes)
				svccommon.HookOnStart(lc, web.Setup)
			}),
		),
		svccommon.Component(
			"webplatform",
			webplatform.Configs,
			fx.Provide(webplatform.New),
			fx.Invoke(func(lc fx.Lifecycle, web *webplatform.Svc) {
				svccommon.HookOnStart(lc, web.Start)
				svccommon.HookOnStop(lc, web.Stop)
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
		fx.Invoke(triggersgrpcsvc.Init),
		fx.Invoke(varsgrpcsvc.Init),
		svccommon.Component(
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
		svccommon.Component("authloginhttpsvc", authloginhttpsvc.Configs, fx.Invoke(authloginhttpsvc.Init)),
		fx.Invoke(func(muxes *muxes.Muxes, h integrationsweb.Handler) {
			muxes.NoAuth.Handle("GET /i/{$}", &h)
		}),
		fx.Invoke(dashboardsvc.Init),
		fx.Invoke(oauth.InitWebhook),
		fx.Invoke(connectionsinitsvc.Init),
		svccommon.Component(
			"webhooks",
			configset.Empty,
			fx.Provide(webhookssvc.New),
			fx.Invoke(func(lc fx.Lifecycle, w *webhookssvc.Service, muxes *muxes.Muxes) {
				svccommon.HookSimpleOnStart(lc, func() { w.Start(muxes) })
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
		svccommon.Component(
			"banner",
			bannerConfigs,
			fx.Invoke(func(cfg *bannerConfig, lc fx.Lifecycle, z *zap.Logger, httpsvc httpsvc.Svc, tclient temporalclient.Client, wp *webplatform.Svc) {
				svccommon.HookSimpleOnStart(lc, func() {
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
	)
}

type RunOptions struct {
	Mode           configset.Mode
	Silent         bool          // No logs at all
	TemporalClient client.Client // use this instead of creating a new temporal client.
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

	var svcs sdkservices.Services = &sdkservices.ServicesStruct{}

	return append(opts, fx.Populate(svcs))
}

func StartDB(ctx context.Context, cfg *svccommon.Config, ropt RunOptions) (db.DB, error) {
	svccommon.SetMode(ropt.Mode)

	var db db.DB

	if err := fx.New(
		fx.Supply(cfg),
		svccommon.LoggerFxOpt(ropt.Silent),
		DBFxOpt(),
		fx.Populate(&db),
	).Start(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
