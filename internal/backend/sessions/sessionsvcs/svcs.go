package sessionsvcs

import (
	"go.uber.org/fx"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authtokens"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/externalclient"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/internal/backend/workflowexecutor"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

type Svcs struct {
	fx.In

	DB db.DB

	Builds       sdkservices.Builds
	Connections  sdkservices.Connections
	Deployments  sdkservices.Deployments
	Events       sdkservices.Events
	Integrations sdkservices.Integrations
	Projects     sdkservices.Projects
	Runtimes     sdkservices.Runtimes
	Store        sdkservices.Store
	Temporal     temporalclient.Client
	Tokens       authtokens.Tokens
	Triggers     sdkservices.Triggers
	Vars         sdkservices.Vars

	WorkflowExecutor workflowexecutor.WorkflowExecutor
	ExternalClient   externalclient.ExternalClient
}
