package proto

import (
	"fmt"

	"github.com/bufbuild/protovalidate-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	applyv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/apply/v1"
	authv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/auth/v1"
	buildsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/builds/v1"
	connectionsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/connections/v1"
	deploymentsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/deployments/v1"
	dispatcherv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/dispatcher/v1"
	envsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/envs/v1"
	eventsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/events/v1"
	integration_providerv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/integration_provider/v1"
	integration_registryv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/integration_registry/v1"
	integrationsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/integrations/v1"
	modulev1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/module/v1"
	oauthv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/oauth/v1"
	programv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/program/v1"
	projectsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/projects/v1"
	runtimesv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runtimes/v1"
	secretsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/secrets/v1"
	sessionsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
	storev1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/store/v1"
	triggersv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/triggers/v1"
	usersv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/users/v1"
	valuesv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/values/v1"
	varsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/vars/v1"
)

func parse(fds []protoreflect.FileDescriptor) func(proto.Message) error {
	var descs []protoreflect.MessageDescriptor

	for _, fd := range fds {
		msgs := fd.Messages()

		for i := 0; i < msgs.Len(); i++ {
			descs = append(descs, msgs.Get(i))
		}
	}

	v, err := protovalidate.New(
		protovalidate.WithDescriptors(descs...),
		protovalidate.WithDisableLazy(true),
	)
	if err != nil {
		panic(fmt.Errorf("protovalidate.New: %w", err))
	}

	return v.Validate
}

var fds = []protoreflect.FileDescriptor{
	applyv1.File_autokitteh_apply_v1_svc_proto,
	authv1.File_autokitteh_auth_v1_svc_proto,
	buildsv1.File_autokitteh_builds_v1_build_proto,
	buildsv1.File_autokitteh_builds_v1_svc_proto,
	connectionsv1.File_autokitteh_connections_v1_connection_proto,
	connectionsv1.File_autokitteh_connections_v1_svc_proto,
	deploymentsv1.File_autokitteh_deployments_v1_deployment_proto,
	deploymentsv1.File_autokitteh_deployments_v1_svc_proto,
	dispatcherv1.File_autokitteh_dispatcher_v1_svc_proto,
	envsv1.File_autokitteh_envs_v1_env_proto,
	envsv1.File_autokitteh_envs_v1_svc_proto,
	eventsv1.File_autokitteh_events_v1_event_proto,
	eventsv1.File_autokitteh_events_v1_svc_proto,
	integration_providerv1.File_autokitteh_integration_provider_v1_integration_proto,
	integration_providerv1.File_autokitteh_integration_provider_v1_svc_proto,
	integration_registryv1.File_autokitteh_integration_registry_v1_integration_proto,
	integration_registryv1.File_autokitteh_integration_registry_v1_svc_proto,
	integrationsv1.File_autokitteh_integrations_v1_integration_proto,
	integrationsv1.File_autokitteh_integrations_v1_svc_proto,
	modulev1.File_autokitteh_module_v1_module_proto,
	oauthv1.File_autokitteh_oauth_v1_svc_proto,
	programv1.File_autokitteh_program_v1_program_proto,
	projectsv1.File_autokitteh_projects_v1_project_proto,
	projectsv1.File_autokitteh_projects_v1_svc_proto,
	runtimesv1.File_autokitteh_runtimes_v1_build_proto,
	runtimesv1.File_autokitteh_runtimes_v1_runtime_proto,
	runtimesv1.File_autokitteh_runtimes_v1_svc_proto,
	secretsv1.File_autokitteh_secrets_v1_svc_proto,
	sessionsv1.File_autokitteh_sessions_v1_session_proto,
	sessionsv1.File_autokitteh_sessions_v1_svc_proto,
	storev1.File_autokitteh_store_v1_svc_proto,
	triggersv1.File_autokitteh_triggers_v1_svc_proto,
	triggersv1.File_autokitteh_triggers_v1_trigger_proto,
	usersv1.File_autokitteh_users_v1_user_proto,
	valuesv1.File_autokitteh_values_v1_values_proto,
	varsv1.File_autokitteh_vars_v1_svc_proto,
	varsv1.File_autokitteh_vars_v1_var_proto,
}

var Validate = parse(fds)
