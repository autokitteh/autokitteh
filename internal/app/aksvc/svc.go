package aksvc

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/fatih/color"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	temporalclient "go.temporal.io/sdk/client"
	"google.golang.org/grpc"

	"github.com/autokitteh/autokitteh/api/gen/openapi"
	webdashboard "github.com/autokitteh/autokitteh/web/dashboard"

	"github.com/autokitteh/autokitteh/internal/app/accountsstoregrpcsvc"
	"github.com/autokitteh/autokitteh/internal/app/croneventsrcsvc"
	"github.com/autokitteh/autokitteh/internal/app/dashboardsvc"
	"github.com/autokitteh/autokitteh/internal/app/eventsrcsstoregrpcsvc"
	"github.com/autokitteh/autokitteh/internal/app/eventsstoregrpcsvc"
	"github.com/autokitteh/autokitteh/internal/app/fseventsrcsvc"
	"github.com/autokitteh/autokitteh/internal/app/githubeventsrcsvc"
	"github.com/autokitteh/autokitteh/internal/app/googleoauthsvc"
	"github.com/autokitteh/autokitteh/internal/app/httpeventsrcsvc"
	"github.com/autokitteh/autokitteh/internal/app/langgrpcsvc"
	"github.com/autokitteh/autokitteh/internal/app/langrungrpcsvc"
	"github.com/autokitteh/autokitteh/internal/app/pluginsgrpcsvc"
	"github.com/autokitteh/autokitteh/internal/app/pluginsreggrpcsvc"
	"github.com/autokitteh/autokitteh/internal/app/projectsstoregrpcsvc"
	"github.com/autokitteh/autokitteh/internal/app/secretsstoregrpcsvc"
	"github.com/autokitteh/autokitteh/internal/app/slackeventsrcsvc"
	"github.com/autokitteh/autokitteh/internal/app/statestoregrpcsvc"

	"github.com/autokitteh/autokitteh/internal/pkg/accountsstore"
	"github.com/autokitteh/autokitteh/internal/pkg/accountsstore/accountsstorefactory"
	"github.com/autokitteh/autokitteh/internal/pkg/akprocs"
	"github.com/autokitteh/autokitteh/internal/pkg/credsstore"
	"github.com/autokitteh/autokitteh/internal/pkg/events"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsrcsstore"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsrcsstore/eventsrcsstorefactory"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsstore"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsstore/eventsstorefactory"
	"github.com/autokitteh/autokitteh/internal/pkg/fseventsrc"
	"github.com/autokitteh/autokitteh/internal/pkg/githubinstalls"
	"github.com/autokitteh/autokitteh/internal/pkg/httpeventsrc"
	"github.com/autokitteh/autokitteh/internal/pkg/lang"
	"github.com/autokitteh/autokitteh/internal/pkg/lang/langrun"
	"github.com/autokitteh/autokitteh/internal/pkg/lang/langrun/locallangrun"
	"github.com/autokitteh/autokitteh/internal/pkg/lang/langtools"
	"github.com/autokitteh/autokitteh/internal/pkg/manifest"
	"github.com/autokitteh/autokitteh/internal/pkg/plugin"
	"github.com/autokitteh/autokitteh/internal/pkg/plugins/internalplugins"
	"github.com/autokitteh/autokitteh/internal/pkg/pluginsreg"
	"github.com/autokitteh/autokitteh/internal/pkg/programs"
	"github.com/autokitteh/autokitteh/internal/pkg/projectsstore"
	"github.com/autokitteh/autokitteh/internal/pkg/projectsstore/projectsstorefactory"
	"github.com/autokitteh/autokitteh/internal/pkg/secretsstore"
	"github.com/autokitteh/autokitteh/internal/pkg/sessions"
	"github.com/autokitteh/autokitteh/internal/pkg/statestore"
	"github.com/autokitteh/autokitteh/internal/pkg/statestore/statestorefactory"
	"github.com/autokitteh/autokitteh/sdk/api/apiplugin"
	"github.com/autokitteh/autokitteh/sdk/api/apiprogram"

	"github.com/autokitteh/autokitteh/assets"
	"github.com/autokitteh/autokitteh/pkg/idgen"
	"github.com/autokitteh/autokitteh/pkg/initd"
	"github.com/autokitteh/autokitteh/pkg/kvstore"
	L "github.com/autokitteh/autokitteh/pkg/l"
	"github.com/autokitteh/autokitteh/pkg/pkvstore"
	"github.com/autokitteh/autokitteh/pkg/svc"
)

var version = "dev"

//go:embed hello.txt
var hello string

var helloTemplate *template.Template

func init() {
	var err error
	if helloTemplate, err = template.New("hello").Parse(hello); err != nil {
		panic(err)
	}
}

// Boxing to distinguish for svc.
type GRPCLangCatalog struct{ lang.Catalog }
type LocalLangCatalog struct{ lang.Catalog }

var SvcOpts = []svc.OptFunc{
	svc.WithConfig(&Config{}),
	svc.WithHTTP(true),
	svc.WithGRPC(true),
	svc.WithComponent(
		svc.Component{
			Name:     "testidgen",
			Disabled: true,
			Init: func() {
				idgen.New = idgen.NewSequentialPerPrefix(0)
			},
		},
		TemporaliteComponent,
		svc.Component{
			Name: "temporalclient",
			Init: func(l L.L, cfg *Config) (temporalclient.Client, error) {
				return temporalclient.NewClient(temporalclient.Options{
					HostPort:  cfg.Temporal.HostPort,
					Namespace: cfg.Temporal.Namespace,
					Logger:    L.Silent{L: l},
				})
			},
		},
		svc.Component{
			Name: "utilitystore",
			Init: func(ctx context.Context, l L.L, cfg *Config) (kvstore.Store, error) {
				return kvstore.Factory{Name: "utility"}.Open(ctx, l, &cfg.UtilityStore)
			},
			Setup: func(ctx context.Context, s kvstore.Store) error {
				return s.Setup(ctx)
			},
		},
		svc.Component{
			Name: "credsstore",
			Init: func(ctx context.Context, l L.L, cfg *Config) *credsstore.Store {
				return &credsstore.Store{Store: pkvstore.Factory{Name: "creds"}.MustOpen(ctx, l, &cfg.CredsStore)}
			},
			Setup: func(ctx context.Context, s *credsstore.Store) error {
				return s.Setup(ctx)
			},
		},
		svc.Component{
			Name: "secretsstore",
			Init: func(ctx context.Context, l L.L, cfg *Config) (*secretsstore.Store, error) {
				pkv, err := pkvstore.Factory{Name: "secrets"}.Open(ctx, l, &cfg.SecretsStore)
				if err != nil {
					return nil, fmt.Errorf("open: %w", err)
				}

				return &secretsstore.Store{Store: pkv}, nil
			},
			Setup: func(ctx context.Context, s *secretsstore.Store) error {
				return s.Setup(ctx)
			},
		},
		svc.Component{
			Name: "secretsstoregrpcsvc",
			Start: func(ctx context.Context, l L.L, store *secretsstore.Store, srv *grpc.Server, gw *runtime.ServeMux) {
				(&secretsstoregrpcsvc.Svc{
					Store: store,
					L:     L.N(l),
				}).Register(ctx, srv, gw)
			},
		},
		svc.Component{
			Name: "statestore",
			Init: func(ctx context.Context, l L.L, cfg *Config) (statestore.Store, error) {
				return statestorefactory.Open(ctx, l, &cfg.StateStore)
			},
			Setup: func(ctx context.Context, s statestore.Store) error {
				return s.Setup(ctx)
			},
		},
		svc.Component{
			Name: "statestoregrpcsvc",
			Start: func(ctx context.Context, l L.L, store statestore.Store, srv *grpc.Server, gw *runtime.ServeMux) {
				(&statestoregrpcsvc.Svc{
					Store: store,
					L:     L.N(l),
				}).Register(ctx, srv, gw)
			},
		},
		svc.Component{
			Name: "accountsstore",
			Init: func(ctx context.Context, l L.L, cfg *Config) (accountsstore.Store, error) {
				return accountsstorefactory.Open(ctx, l, &cfg.AccountsStore)
			},
			Setup: func(ctx context.Context, as accountsstore.Store) error {
				return as.Setup(ctx)
			},
		},
		svc.Component{
			Name: "accountsstoregrpcsvc",
			Start: func(ctx context.Context, l L.L, accounts accountsstore.Store, srv *grpc.Server, gw *runtime.ServeMux) {
				(&accountsstoregrpcsvc.Svc{
					Store: accounts,
					L:     L.N(l),
				}).Register(ctx, srv, gw)
			},
		},
		svc.Component{
			Name: "projectsstore",
			Init: func(ctx context.Context, l L.L, cfg *Config, accounts accountsstore.Store) (projectsstore.Store, error) {
				return projectsstorefactory.Open(ctx, l, &cfg.ProjectsStore, accounts)
			},
			Setup: func(ctx context.Context, ps projectsstore.Store) error {
				return ps.Setup(ctx)
			},
		},
		svc.Component{
			Name: "eventsstore",
			Init: func(ctx context.Context, l L.L, cfg *Config) (eventsstore.Store, error) {
				return eventsstorefactory.Open(ctx, l, &cfg.EventsStore)
			},
			Setup: func(ctx context.Context, as eventsstore.Store) error {
				return as.Setup(ctx)
			},
		},
		svc.Component{
			Name: "eventsrcsstore",
			Init: func(ctx context.Context, l L.L, cfg *Config) (eventsrcsstore.Store, error) {
				return eventsrcsstorefactory.Open(ctx, l, &cfg.EventSourcesStore)
			},
			Setup: func(ctx context.Context, es eventsrcsstore.Store) error {
				return es.Setup(ctx)
			},
		},
		svc.Component{
			Name: "eventsrcsstoregrpcsvc",
			Start: func(ctx context.Context, l L.L, eventsrcs eventsrcsstore.Store, srv *grpc.Server) {
				(&eventsrcsstoregrpcsvc.Svc{Store: eventsrcs, L: L.N(l)}).Register(ctx, srv)
			},
		},
		svc.Component{
			Name: "eventsstoregrpcsvc",
			Start: func(ctx context.Context, l L.L, eventsStore eventsstore.Store, events *events.Events, srv *grpc.Server, gw *runtime.ServeMux) {
				(&eventsstoregrpcsvc.Svc{Events: events, Store: eventsStore, L: L.N(l)}).Register(ctx, srv, gw)
			},
		},
		svc.Component{
			Name: "projectsstoregrpcsvc",
			Start: func(ctx context.Context, l L.L, projects projectsstore.Store, srv *grpc.Server, gw *runtime.ServeMux) {
				(&projectsstoregrpcsvc.Svc{Store: projects, L: L.N(l)}).Register(ctx, srv, gw)
			},
		},
		svc.Component{
			Name: "local_catalog",
			Init: func(cfg *Config, l L.L) LocalLangCatalog {
				if cfg.CatalogPermissive {
					l.Warn("using permissive catalog")
					return LocalLangCatalog{Catalog: langtools.PermissiveCatalog}
				}

				return LocalLangCatalog{Catalog: langtools.DeterministicCatalog}
			},
		},
		svc.Component{
			Name: "langgrpcsvc",
			Init: func(ctx context.Context, l L.L, srv *grpc.Server, cat LocalLangCatalog) *langgrpcsvc.Svc {
				return &langgrpcsvc.Svc{L: L.N(l), Catalog: cat}
			},
			Start: func(ctx context.Context, l L.L, svc *langgrpcsvc.Svc, srv *grpc.Server, gw *runtime.ServeMux) {
				svc.Register(ctx, srv, gw)
			},
		},
		svc.Component{
			Name: "langrun",
			Init: func(ctx context.Context, l L.L, cat LocalLangCatalog) langrun.Runs {
				return locallangrun.NewRuns(l, cat, nil, nil)
			},
		},
		svc.Component{
			Name: "langrungrpcsvc",
			Init: func(ctx context.Context, l L.L, runs langrun.Runs) *langrungrpcsvc.Svc {
				return &langrungrpcsvc.Svc{L: L.N(l), Runs: runs}
			},
			Start: func(ctx context.Context, l L.L, runs langrun.Runs, svc *langrungrpcsvc.Svc, srv *grpc.Server, gw *runtime.ServeMux) {
				svc.Register(ctx, srv, gw)
			},
		},
		svc.Component{
			Name: "grpccatalog",
			Init: func(ctx context.Context, l L.L, langsvc *langgrpcsvc.Svc, langrunsvc *langrungrpcsvc.Svc) (GRPCLangCatalog, error) {
				// HACK: Use GRPC catalog with proxying to local makes lang run summary work.
				cat, err := langtools.NewGRPCCatalog(
					ctx,
					l,
					&langgrpcsvc.LocalClient{Server: langsvc},
					&langrungrpcsvc.LocalClient{Server: langrunsvc},
				)
				if err != nil {
					return GRPCLangCatalog{}, err
				}

				langs := cat.List()
				l.Debug("supported languages", "langs", langs)

				return GRPCLangCatalog{Catalog: cat}, nil
			},
		},
		svc.Component{
			Name:     "githubinstalls",
			Disabled: true,
			Init: func(ctx context.Context, l L.L, cfg *Config) *githubinstalls.Installs {
				return githubinstalls.New(
					cfg.GithubInstalls,
					kvstore.Factory{Name: "githubinstallations"}.MustOpen(ctx, l, &cfg.GithubInstallsStore),
				)
			},
			Setup: func(ctx context.Context, i *githubinstalls.Installs) error {
				return i.Store.Setup(ctx)
			},
		},
		svc.Component{
			Name: "programs",
			Init: func(ctx context.Context, cfg *Config, l L.L, cat GRPCLangCatalog, github *githubinstalls.Installs) *programs.Programs {
				return &programs.Programs{
					L:       L.N(l),
					Catalog: cat,
					PathRewriters: []programs.PathRewriterFunc{
						programs.GithubPathRewriter,
					},
					CommonLoaders: map[string]programs.LoaderFunc{
						"$internal": func(ctx context.Context, path *apiprogram.Path) ([]byte, error) {
							if path.Scheme() != "$internal" {
								return nil, fmt.Errorf("invalid scheme %q", path.Scheme())
							}

							p := filepath.Clean(path.Path())

							if p == "" || strings.Contains(p, "..") || p[0] == os.PathSeparator {
								return nil, fmt.Errorf("invalid path %q -> %q", path.Path(), p)

							}

							return assets.FS.ReadFile(filepath.Join("internal", p))
						},
						"github": programs.NewGithubLoader(l.Named("githubloader"), github.GetClient),
					},
				}
			},
		},
		svc.Component{
			Name: "pluginsreg",
			Init: func(ctx context.Context, l L.L, cfg *Config) *pluginsreg.Registry {
				reg := &pluginsreg.Registry{
					L:     L.N(l),
					Procs: &akprocs.Procs{Config: cfg.PluginsRegProcs, L: L.N(l)},
					Store: pkvstore.Factory{Name: "pluginsreg"}.MustOpen(ctx, l, &cfg.CredsStore),
				}

				internalplugins.RegisterAll(reg.RegisterInternalPlugin)

				return reg
			},
			Setup: func(ctx context.Context, s *pluginsreg.Registry) error {
				return s.Store.Setup(ctx)
			},
			Start: func(r *mux.Router, reg *pluginsreg.Registry) {
				reg.Procs.Register(r.PathPrefix("/pluginsreg/").Subrouter())
			},
		},
		svc.Component{
			Name: "pluginsreggrpcsvc",
			Init: func(l L.L, reg *pluginsreg.Registry) *pluginsreggrpcsvc.Svc {
				return &pluginsreggrpcsvc.Svc{L: L.N(l), Registry: reg}
			},
			Start: func(srv *grpc.Server, svc *pluginsreggrpcsvc.Svc) {
				svc.RegisterServer(srv)
			},
		},
		svc.Component{
			Name: "pluginsgrpcsvc",
			Init: func(l L.L) *pluginsgrpcsvc.Svc {
				return &pluginsgrpcsvc.Svc{
					L:       L.N(l),
					Plugins: map[apiplugin.PluginID]plugin.Plugin{},
				}
			},
			Start: func(srv *grpc.Server, svc *pluginsgrpcsvc.Svc) {
				svc.Register(srv)
			},
		},
		svc.Component{
			Name: "sessions",
			Init: func(
				ctx context.Context,
				l L.L,
				temporal temporalclient.Client,
				programs *programs.Programs,
				eventsStore eventsstore.Store,
				stateStore statestore.Store,
				secretsStore *secretsstore.Store,
				credsStore *credsstore.Store,
				pluginsReg *pluginsreg.Registry,
			) *sessions.Sessions {
				es := &sessions.Sessions{
					Temporal:    temporal,
					L:           L.N(l),
					Programs:    programs,
					Plugins:     pluginsReg,
					EventsStore: eventsStore,
					StateStore:  stateStore,
					GetSecret:   secretsStore.Get,
					GetCreds:    credsStore.Get,
				}

				es.Init()

				return es
			},
			Start: func(s *sessions.Sessions) error { return s.Start() },
		},
		svc.Component{
			Name: "events",
			Init: func(
				l L.L,
				temporal temporalclient.Client,
				accountsStore accountsstore.Store,
				projectsStore projectsstore.Store,
				eventsStore eventsstore.Store,
				eventsrcsStore eventsrcsstore.Store,
				sessions *sessions.Sessions,
			) *events.Events {
				es := &events.Events{
					Temporal: temporal,
					L:        L.N(l),
					Run:      sessions.Run,
					Stores: events.Stores{
						Events:       eventsStore,
						EventSources: eventsrcsStore,
						Accounts:     accountsStore,
						Projects:     projectsStore,
					},
				}

				es.Init()

				return es
			},
			Start: func(es *events.Events) error {
				return es.Start()
			},
		},
		svc.Component{
			Name:     "fseventsrc",
			Disabled: true,
			Init: func(ctx context.Context, l L.L, cfg *Config, srcs eventsrcsstore.Store, events *events.Events) *fseventsrc.FSEventSource {
				return &fseventsrc.FSEventSource{
					Config:       cfg.FSEventSource,
					L:            L.N(l),
					Events:       events,
					EventSources: srcs,
				}
			},
			Start: func(ctx context.Context, es *fseventsrc.FSEventSource) error {
				return es.Start(ctx)
			},
		},
		svc.Component{
			Name: "googleoauthssvc",
			Init: func(ctx context.Context, l L.L, cfg *Config, creds *credsstore.Store, ustore kvstore.Store) *googleoauthsvc.Svc {
				return &googleoauthsvc.Svc{
					L:               L.N(l),
					CredsStore:      creds,
					OAuthStateStore: &kvstore.StoreWithKeyPrefix{Store: ustore, Prefix: "googleoauthsvc_state_"},
					Config:          cfg.GoogleOAuthSvc,
				}
			},
			Start: func(ctx context.Context, svc *googleoauthsvc.Svc, r *mux.Router) {
				svc.Register(r)
			},
		},
		svc.Component{
			Name:     "githubeventsrcsvc",
			Disabled: true,
			Init: func(ctx context.Context, l L.L, cfg *Config, creds *credsstore.Store, installs *githubinstalls.Installs, srcs eventsrcsstore.Store, events *events.Events) *githubeventsrcsvc.Svc {
				return &githubeventsrcsvc.Svc{
					L:            L.N(l),
					EventSources: srcs,
					Events:       events,
					Installs:     installs,
					Config:       cfg.GithubEventSource,
				}
			},
			Start: func(ctx context.Context, src *githubeventsrcsvc.Svc, r *mux.Router, srv *grpc.Server, gw *runtime.ServeMux) {
				src.Register(ctx, srv, gw, r)
			},
		},
		svc.Component{
			Name:     "slackeventsrcsvc",
			Disabled: true,
			Init: func(ctx context.Context, l L.L, cfg *Config, creds *credsstore.Store, srcs eventsrcsstore.Store, events *events.Events) *slackeventsrcsvc.Svc {
				return &slackeventsrcsvc.Svc{
					L:            L.N(l),
					EventSources: srcs,
					Events:       events,
					CredsStore:   creds,
					Config:       cfg.SlackEventSource,
				}
			},
			Start: func(ctx context.Context, src *slackeventsrcsvc.Svc, srv *grpc.Server, gw *runtime.ServeMux, r *mux.Router) error {
				return src.Start(ctx, srv, gw, r)
			},
		},
		svc.Component{
			Name: "fseventsrcsvc",
			Start: func(ctx context.Context, l L.L, src *fseventsrc.FSEventSource, srv *grpc.Server, gw *runtime.ServeMux) {
				if src != nil {
					(&fseventsrcsvc.Svc{L: L.N(l), Src: src}).Register(ctx, srv, gw)
				}
			},
		},
		svc.Component{
			Name: "httpeventsrc",
			Init: func(ctx context.Context, l L.L, cfg *Config, events *events.Events, srcs eventsrcsstore.Store) (*httpeventsrc.HTTPEventSource, error) {
				return &httpeventsrc.HTTPEventSource{
					L:            L.N(l),
					Events:       events,
					EventSources: srcs,
					Config:       cfg.HTTPEventSource,
					Prefix:       "/httpsrc/",
				}, nil
			},
		},
		svc.Component{
			Name: "httpeventsrcsvc",
			Start: func(ctx context.Context, l L.L, src *httpeventsrc.HTTPEventSource, r *mux.Router, srv *grpc.Server, gw *runtime.ServeMux) {
				if src != nil {
					(&httpeventsrcsvc.Svc{L: L.N(l), Src: src}).Register(ctx, srv, gw, r.PathPrefix("/httpsrc/"))
				}
			},
		},
		svc.Component{
			Name:     "croneventsrcsvc",
			Disabled: true,
			Init: func(cfg *Config, l L.L, srcs eventsrcsstore.Store, ustore kvstore.Store, events *events.Events) *croneventsrcsvc.Svc {
				return &croneventsrcsvc.Svc{
					L:            L.N(l),
					Config:       cfg.CronEventSource,
					Events:       events,
					EventSources: srcs,
					StateStore:   &kvstore.StoreWithKeyPrefix{Store: ustore, Prefix: "croneventsrc_"},
				}
			},
			Start: func(ctx context.Context, svc *croneventsrcsvc.Svc, srv *grpc.Server, gw *runtime.ServeMux) {
				svc.Register(ctx, srv, gw)
				svc.Start()
			},
		},
		svc.Component{
			Name:     "defaults",
			Disabled: true,
			Ready: func(p *programs.Programs) {
				p.SetCommonLoader("fs", programs.NewFSLoader(os.DirFS("."), "."))
			},
		},
		svc.Component{
			Name: "openapi",
			Start: func(r *mux.Router) {
				r.
					PathPrefix("/openapi").
					Handler(
						http.StripPrefix(
							"/openapi",
							http.FileServer(http.FS(openapi.FS)),
						),
					)
			},
		},
		svc.Component{
			Name: "dash",
			Start: func(cfg *Config, r *mux.Router, l L.L) {
				fs := http.FS(webdashboard.FS)
				if !cfg.EmbeddedDash {
					l.Info("serving dashboard from filesystem")
					fs = http.Dir("web/dashboard/build")
				} else {
					l.Info("serving dashboard from filesystem")
				}

				r.
					PathPrefix("/dash/").
					Handler(
						http.StripPrefix("/dash/", http.FileServer(fs)),
					)
			},
		},
		svc.Component{
			Name: "dashboard",
			Start: func(
				cfg *Config,
				r *mux.Router,
				eventsStore eventsstore.Store,
				projectsStore projectsstore.Store,
				eventSrcsStore eventsrcsstore.Store,
				stateStore statestore.Store,
				secretsStore *secretsstore.Store,
			) {
				(&dashboardsvc.Svc{
					Config:            cfg.Dashboard,
					EventsStore:       eventsStore,
					ProjectsStore:     projectsStore,
					EventSourcesStore: eventSrcsStore,
					StateStore:        stateStore,
					SecretsStore:      secretsStore,
				}).Register(r)
			},
		},
		svc.Component{
			Name: "initd",
			Ready: func(cfg *Config, l L.L, svcCfg svc.SvcCfg) error {
				return (&initd.Initd{
					Config: cfg.Initd,
					L:      L.N(l),
					Env: map[string]string{
						"AK_GRPC_ADDR": fmt.Sprintf("127.0.0.1:%d", svcCfg.GRPC.Port),
						"AK_HTTP_ADDR": fmt.Sprintf("http://127.0.0.1:%d", svcCfg.HTTP.Port),
					},
				}).Start()
			},
		},
		svc.Component{
			Name: "hello",
			Ready: func(
				ctx context.Context,
				svcCfg *svc.SvcCfg,
				accounts accountsstore.Store,
				projects projectsstore.Store,
				eventsrcs eventsrcsstore.Store,
				httpeventssrc *httpeventsrc.HTTPEventSource,
				fseventssrc *fseventsrc.FSEventSource,
			) error {
				httpPort, grpcPort := fmt.Sprintf("%d", svcCfg.HTTP.Port), fmt.Sprintf("%d", svcCfg.GRPC.Port)
				if !svcCfg.HTTP.Enabled {
					httpPort = "disabled"
				}

				if !svcCfg.GRPC.Enabled {
					grpcPort = "disabled"
				}

				valColor := color.New(color.FgCyan).SprintFunc()

				data := struct {
					Version                                                string
					PID                                                    string
					HTTPPort, GRPCPort                                     string
					Extra0, Extra1, Extra2, Extra3, Extra4, Extra5, Extra6 string
				}{
					Version:  version,
					PID:      valColor(fmt.Sprintf("%d", os.Getpid())),
					HTTPPort: valColor(httpPort),
					GRPCPort: valColor(grpcPort),
				}

				if err := helloTemplate.Execute(os.Stdout, data); err != nil {
					return fmt.Errorf("hello template render: %w", err)
				}

				return nil
			},
		},
		svc.Component{
			Name: "initmanifest",
			Init: func(l L.L, cfg *Config) (actions manifest.Actions, _ error) {
				for _, initpath := range cfg.InitPaths {
					m, err := manifest.ManifestFromPath(initpath)
					if err != nil {
						return nil, fmt.Errorf("%q: %w", initpath, err)
					}

					if m == nil {
						return nil, fmt.Errorf("%q: not found", initpath)
					}

					l.Debug("loaded init manifest", "path", initpath, "manifest", m)

					as, err := m.Compile()
					if err != nil {
						return nil, fmt.Errorf("%q: %w", initpath, err)
					}

					l.Debug("compiled init manifest", "path", initpath, "num_actions", len(as))

					actions = append(actions, as...)
				}

				return
			},
			Setup: func(
				ctx context.Context,
				l L.L,
				eventsrcs eventsrcsstore.Store,
				projects projectsstore.Store,
				accounts accountsstore.Store,
				plugins *pluginsreg.Registry,
				actions manifest.Actions,
			) error {
				if len(actions) == 0 {
					l.Debug("no init actions")
					return nil
				}

				l.Info("+ applying init actions")

				env := &manifest.Env{
					EventSources: eventsrcs,
					Projects:     projects,
					Accounts:     accounts,
					Plugins:      plugins,
				}

				log, err := env.Apply(ctx, actions)

				for _, log1 := range log {
					l.Info(fmt.Sprintf("| %s", log1))
				}

				l.Info(fmt.Sprintf("+ %d initializations applied", len(log)))

				return err
			},
		},
		svc.Component{
			Name: "grpcgw",
			Init: func() *runtime.ServeMux { return runtime.NewServeMux() },
			Start: func(mux *runtime.ServeMux, r *mux.Router) {
				r.PathPrefix("/api/").Handler(mux)
			},
		},
	),
}
