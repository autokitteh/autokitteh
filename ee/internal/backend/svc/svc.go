package svc

import (
	"context"
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/ee/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/ee/internal/backend/auth/authgrpcsvc"
	"go.autokitteh.dev/autokitteh/ee/internal/backend/auth/authhttpmiddleware"
	"go.autokitteh.dev/autokitteh/ee/internal/backend/auth/authjwttokens"
	"go.autokitteh.dev/autokitteh/ee/internal/backend/auth/authloginhttpsvc"
	"go.autokitteh.dev/autokitteh/ee/internal/backend/auth/authsessions"
	"go.autokitteh.dev/autokitteh/ee/internal/backend/auth/authsvc"
	"go.autokitteh.dev/autokitteh/ee/internal/backend/db"
	"go.autokitteh.dev/autokitteh/ee/internal/backend/db/dbfactory"
	"go.autokitteh.dev/autokitteh/ee/internal/backend/envsauth"
	"go.autokitteh.dev/autokitteh/ee/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/ee/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/ee/internal/backend/orgs"
	"go.autokitteh.dev/autokitteh/ee/internal/backend/orgsgrpcsvc"
	"go.autokitteh.dev/autokitteh/ee/internal/backend/projectsauth"
	"go.autokitteh.dev/autokitteh/ee/internal/backend/users"
	"go.autokitteh.dev/autokitteh/ee/internal/backend/usersgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/basesvc"
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/httpsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/integrations"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/auth/v1/authv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/builds/v1/buildsv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/connections/v1/connectionsv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/deployments/v1/deploymentsv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/dispatcher/v1/dispatcherv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/envs/v1/envsv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/events/v1/eventsv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/integrations/v1/integrationsv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/oauth/v1/oauthv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/orgs/v1/orgsv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/projects/v1/projectsv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runtimes/v1/runtimesv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/secrets/v1/secretsv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1/sessionsv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/users/v1/usersv1connect"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

func makeFxOpts(cfg *basesvc.Config) []fx.Option {
	return []fx.Option{
		basesvc.Component(
			"db",
			dbfactory.Configs,
			fx.Provide(dbfactory.New),
			fx.Invoke(func(lc fx.Lifecycle, z *zap.Logger, db db.DB) {
				basesvc.HookOnStart(lc, db.Connect)
				basesvc.HookOnStart(lc, db.Setup)
			}),
		),

		basesvc.Component("envsauth", configset.Empty, fx.Decorate(envsauth.Wrap)),
		basesvc.Component("projectsauth", configset.Empty, fx.Decorate(projectsauth.Wrap)),
		basesvc.Component("orgs", configset.Empty, fx.Provide(orgs.New)),
		basesvc.Component("users", configset.Empty, fx.Provide(users.New)),
		basesvc.Component("auth", configset.Empty, fx.Provide(authsvc.New)),

		basesvc.Component("authjwttokens", authjwttokens.Configs, fx.Provide(authjwttokens.New)),
		basesvc.Component("authsessions", authsessions.Configs, fx.Provide(authsessions.New)),
		basesvc.Component("httpauthmiddleware", authhttpmiddleware.Configs, fx.Provide(authhttpmiddleware.New)),
		basesvc.Component(
			"http",
			httpsvc.Configs,
			fx.Provide(func(lc fx.Lifecycle, z *zap.Logger, cfg *httpsvc.Config, wrap authhttpmiddleware.WrapFunc) (*http.ServeMux, muxes.Muxes) {
				mux := httpsvc.New(
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
						mappingsv1connect.MappingsServiceName,
						oauthv1connect.OAuthServiceName,
						orgsv1connect.OrgsServiceName,
						projectsv1connect.ProjectsServiceName,
						runtimesv1connect.RuntimesServiceName,
						secretsv1connect.SecretsServiceName,
						sessionsv1connect.SessionsServiceName,
						usersv1connect.UsersServiceName,
					},
					[]httpsvc.RequestLogExtractor{
						func(r *http.Request) []zap.Field {
							if userID := authcontext.GetAuthnUserID(r.Context()); userID != nil {
								return []zap.Field{zap.String("user_id", userID.String())}
							}

							return nil
						},
					},
				)

				// Replace the original mux with a main mux and also expose the original mux as the no-auth one.
				wrapped := http.NewServeMux()
				mux.Handle("/", wrap(wrapped))
				return wrapped, muxes.Muxes{Auth: wrapped, NoAuth: mux}
			}),
		),
		basesvc.Component("authloginhttpsvc", authloginhttpsvc.Configs, fx.Invoke(authloginhttpsvc.Init)),

		fx.Invoke(authgrpcsvc.Init),
		fx.Invoke(orgsgrpcsvc.Init),
		fx.Invoke(usersgrpcsvc.Init),

		basesvc.Component(
			"integrations",
			configset.Empty,
			fx.Provide(integrations.New(fixtures.AutokittehOrgID)),
			fx.Invoke(func(lc fx.Lifecycle, l *zap.Logger, mux *http.ServeMux, s sdkservices.Secrets, o sdkservices.OAuth, d sdkservices.Dispatcher) {
				basesvc.HookOnStart(lc, func(ctx context.Context) error {
					return integrations.Start(ctx, l, mux, s, o, d)
				})
			}),
		),

		indexOption(),
	}
}

func NewOpts(cfg *basesvc.Config, ropts basesvc.RunOptions) []fx.Option {
	basesvc.SetFXRunOpts(ropts)

	return append(basesvc.NewCommonOpts(cfg, ropts), makeFxOpts(cfg)...)
}

func New(cfg *basesvc.Config, ropts basesvc.RunOptions) *fx.App {
	return fx.New(NewOpts(cfg, ropts)...)
}
