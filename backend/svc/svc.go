package svc

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/backend/basesvc"
	"go.autokitteh.dev/autokitteh/backend/configset"
	"go.autokitteh.dev/autokitteh/backend/httpsvc"
	"go.autokitteh.dev/autokitteh/backend/integrations"
	"go.autokitteh.dev/autokitteh/backend/internal/dispatcher"
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
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

//go:embed index.html
var indexHTML string

// New returns after all services has been created, but not started (no fx.Invoke
// options are called).
func NewOpts(cfg *basesvc.Config, ropts basesvc.RunOptions) []fx.Option {
	basesvc.SetFXRunOpts(ropts)

	fxOpts := []fx.Option{
		basesvc.Component(
			"http",
			httpsvc.Configs,
			fx.Provide(func(lc fx.Lifecycle, z *zap.Logger, cfg *httpsvc.Config) (*http.ServeMux, error) {
				return httpsvc.New(
					lc, z, cfg,
					[]string{
						buildsv1connect.BuildsServiceName,
						connectionsv1connect.ConnectionsServiceName,
						deploymentsv1connect.DeploymentsServiceName,
						dispatcherv1connect.DispatcherServiceName,
						envsv1connect.EnvsServiceName,
						eventsv1connect.EventsServiceName,
						integrationsv1connect.IntegrationsServiceName,
						triggersv1connect.TriggersServiceName,
						oauthv1connect.OAuthServiceName,
						projectsv1connect.ProjectsServiceName,
						runtimesv1connect.RuntimesServiceName,
						secretsv1connect.SecretsServiceName,
						sessionsv1connect.SessionsServiceName,
					},
					nil,
				)
			}),
		),
		basesvc.Component("integrations", configset.Empty, fx.Provide(integrations.New)),
		fx.Invoke(func(lc fx.Lifecycle, l *zap.Logger, mux *http.ServeMux, s sdkservices.Secrets, o sdkservices.OAuth, d dispatcher.Dispatcher) {
			basesvc.HookOnStart(lc, func(ctx context.Context) error {
				return integrations.Start(ctx, l, mux, s, o, d)
			})
		}),
		fx.Invoke(func(z *zap.Logger, mux *http.ServeMux) {
			mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, indexHTML)
			}))
		}),
	}

	return append(basesvc.NewCommonOpts(cfg, ropts), fxOpts...)
}

func New(cfg *basesvc.Config, ropts basesvc.RunOptions) *fx.App {
	return fx.New(NewOpts(cfg, ropts)...)
}
