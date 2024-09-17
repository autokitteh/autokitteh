package proto

import (
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/apply/v1/applyv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/auth/v1/authv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/builds/v1/buildsv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/connections/v1/connectionsv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/deployments/v1/deploymentsv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/dispatcher/v1/dispatcherv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/envs/v1/envsv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/events/v1/eventsv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/integration_provider/v1/integration_providerv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/integration_registry/v1/integration_registryv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/integrations/v1/integrationsv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/oauth/v1/oauthv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/projects/v1/projectsv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runtimes/v1/runtimesv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1/sessionsv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/store/v1/storev1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/triggers/v1/triggersv1connect"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/vars/v1/varsv1connect"
)

var ServiceNames = []string{
	applyv1connect.ApplyServiceName,
	authv1connect.AuthServiceName,
	buildsv1connect.BuildsServiceName,
	connectionsv1connect.ConnectionsServiceName,
	deploymentsv1connect.DeploymentsServiceName,
	dispatcherv1connect.DispatcherServiceName,
	envsv1connect.EnvsServiceName,
	eventsv1connect.EventsServiceName,
	integration_providerv1connect.IntegrationProviderServiceName,
	integration_registryv1connect.IntegrationRegistryServiceName,
	integrationsv1connect.IntegrationsServiceName,
	oauthv1connect.OAuthServiceName,
	projectsv1connect.ProjectsServiceName,
	runtimesv1connect.RuntimesServiceName,
	sessionsv1connect.SessionsServiceName,
	storev1connect.StoreServiceName,
	triggersv1connect.TriggersServiceName,
	varsv1connect.VarsServiceName,
}
