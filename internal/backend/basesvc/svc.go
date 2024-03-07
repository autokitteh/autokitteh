package basesvc

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

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
	"go.autokitteh.dev/autokitteh/internal/backend/integrationsgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/logger"
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
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkruntimessvc"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/web/integrations"
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
	return Component(
		"db",
		dbfactory.Configs,
		fx.Provide(dbfactory.New),
		fx.Invoke(func(lc fx.Lifecycle, z *zap.Logger, db db.DB) {
			HookOnStart(lc, db.Connect)
		}),
	)
}

func makeFxOpts(cfg *Config, opts RunOptions) []fx.Option {
	return []fx.Option{
		fx.Supply(cfg),
		LoggerFxOpt(),
		DBFxOpt(),
		fx.Invoke(func(lc fx.Lifecycle, db db.DB) {
			HookOnStart(lc, db.Setup)
		}),
		Component(
			"temporalclient",
			temporalclient.Configs,
			fx.Provide(temporalclient.New),
			fx.Provide(func(c temporalclient.Client) client.Client { return c.Temporal() }),
			fx.Invoke(func(lc fx.Lifecycle, c temporalclient.Client) {
				HookOnStart(lc, c.Start)
				HookOnStop(lc, c.Stop)
			}),
		),
		Component("secrets", configset.Empty, fx.Provide(secrets.New)),
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

		fx.Provide(func(s fxServices) sdkservices.Services { return &s }),

		fx.Invoke(sdkruntimessvc.Init),
		fx.Invoke(applygrpcsvc.Init),
		fx.Invoke(func(p *projectsgrpcsvc.Server, mux *http.ServeMux) {
			p.Init(mux)
		}),
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
		fx.Invoke(func(mux *http.ServeMux, h integrations.Handler) {
			mux.Handle("/i/", &h)
		}),
		fx.Invoke(func(mux *http.ServeMux, l *zap.Logger, s sdkservices.Services) {
			mux.Handle("/oauth/", oauth.NewWebhook(l, s))
		}),
		fx.Invoke(func(z *zap.Logger, mux *http.ServeMux) {
			srv := http.StripPrefix("/static/", http.FileServer(http.FS(static.RootWebContent)))
			mux.Handle("/static/", srv)
			mux.Handle("/favicon-32x32.png", srv)
			mux.Handle("/favicon-16x16.png", srv)
		}),
		fx.Invoke(func(lc fx.Lifecycle, z *zap.Logger, httpsvc httpsvc.Svc) {
			HookSimpleOnStart(lc, func() {
				sayHello(opts, httpsvc.Addr())
				z.Info("ready")
			})
		}),
	}
}

type RunOptions struct {
	Mode   configset.Mode
	Silent bool // No logs at all
}

func NewCommonOpts(cfg *Config, ropts RunOptions) []fx.Option {
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

func StartDB(ctx context.Context, cfg *Config) (db.DB, error) {
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
