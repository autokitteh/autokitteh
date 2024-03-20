package db

import (
	"context"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Transaction interface {
	DB
	Commit() error

	// Does nothing if already committed.
	Rollback() error
}

func LoggedRollback(z *zap.Logger, tx Transaction) {
	if err := tx.Rollback(); err != nil {
		z.Error("rollback error", zap.Error(err))
	}
}

type DB interface {
	Connect(context.Context) error
	Setup(context.Context) error
	Teardown(context.Context) error

	Debug() DB

	// Begina a transaction.
	Begin(context.Context) (Transaction, error)

	Transaction(context.Context, func(tx DB) error) error

	// -----------------------------------------------------------------------
	// Returns sdkerrors.ErrAlreadyExists if either id or name is duplicate.
	CreateProject(context.Context, sdktypes.Project) error

	UpdateProject(context.Context, sdktypes.Project) error

	// Returns sdkerrors.ErrNotFound if not found.
	GetProjectByID(context.Context, sdktypes.ProjectID) (sdktypes.Project, error)

	// Returns sdkerrors.ErrNotFound if not found.
	GetProjectByName(context.Context, sdktypes.Symbol) (sdktypes.Project, error)

	ListProjects(context.Context) ([]sdktypes.Project, error)

	// Returns nill, nil if no resources are set.
	GetProjectResources(context.Context, sdktypes.ProjectID) (map[string][]byte, error)

	SetProjectResources(context.Context, sdktypes.ProjectID, map[string][]byte) error

	// deletes a project and all its resources
	DeleteProject(context.Context, sdktypes.ProjectID) error

	// -----------------------------------------------------------------------
	// Returns sdkerrors.ErrAlreadyExists if either id or name is duplicate.
	CreateEnv(context.Context, sdktypes.Env) error

	// Returns sdkerrors.ErrNotFound if id is not found.
	GetEnvByID(context.Context, sdktypes.EnvID) (sdktypes.Env, error)

	// Returns sdkerrors.ErrNotFound if name is not found.
	GetEnvByName(context.Context, sdktypes.ProjectID, sdktypes.Symbol) (sdktypes.Env, error)

	ListProjectEnvs(context.Context, sdktypes.ProjectID) ([]sdktypes.Env, error)

	SetEnvVar(context.Context, sdktypes.EnvVar) error

	GetEnvVars(context.Context, sdktypes.EnvID) ([]sdktypes.EnvVar, error)

	// Return sdkerrors.ErrNotFound if var not found.
	RevealEnvVar(context.Context, sdktypes.EnvID, sdktypes.Symbol) (string, error)

	// -----------------------------------------------------------------------
	// This is idempotent.
	SaveEvent(context.Context, sdktypes.Event) error
	GetEventByID(context.Context, sdktypes.EventID) (sdktypes.Event, error)
	ListEvents(context.Context, sdkservices.ListEventsFilter) ([]sdktypes.Event, error)
	AddEventRecord(context.Context, sdktypes.EventRecord) error
	ListEventRecords(context.Context, sdkservices.ListEventRecordsFilter) ([]sdktypes.EventRecord, error)
	GetLatestEventSequence(context.Context) (uint64, error)

	// -----------------------------------------------------------------------
	CreateTrigger(context.Context, sdktypes.Trigger) error
	UpdateTrigger(context.Context, sdktypes.Trigger) error
	GetTrigger(context.Context, sdktypes.TriggerID) (sdktypes.Trigger, error)
	DeleteTrigger(context.Context, sdktypes.TriggerID) error
	ListTriggers(context.Context, sdkservices.ListTriggersFilter) ([]sdktypes.Trigger, error)

	// -----------------------------------------------------------------------
	GetBuild(ctx context.Context, buildID sdktypes.BuildID) (sdktypes.Build, error)
	ListBuilds(ctx context.Context, filter sdkservices.ListBuildsFilter) ([]sdktypes.Build, error)
	GetBuildData(ctx context.Context, id sdktypes.BuildID) ([]byte, error)
	SaveBuild(ctx context.Context, Build sdktypes.Build, data []byte) error
	DeleteBuild(ctx context.Context, buildID sdktypes.BuildID) error

	// -----------------------------------------------------------------------
	CreateConnection(ctx context.Context, conn sdktypes.Connection) error
	UpdateConnection(ctx context.Context, conn sdktypes.Connection) error
	DeleteConnection(ctx context.Context, id sdktypes.ConnectionID) error
	GetConnection(ctx context.Context, id sdktypes.ConnectionID) (sdktypes.Connection, error)
	ListConnections(ctx context.Context, filter sdkservices.ListConnectionsFilter) ([]sdktypes.Connection, error)

	// -----------------------------------------------------------------------
	GetDeployment(ctx context.Context, id sdktypes.DeploymentID) (sdktypes.Deployment, error)

	// Returns deployments in ascending order by creation time.
	ListDeployments(ctx context.Context, filter sdkservices.ListDeploymentsFilter) ([]sdktypes.Deployment, error)
	UpdateDeploymentState(ctx context.Context, id sdktypes.DeploymentID, state sdktypes.DeploymentState) error
	CreateDeployment(ctx context.Context, deployment sdktypes.Deployment) error
	DeleteDeployment(ctx context.Context, deploymentID sdktypes.DeploymentID) error

	// -----------------------------------------------------------------------
	CreateIntegration(ctx context.Context, i sdktypes.Integration) error
	UpdateIntegration(ctx context.Context, i sdktypes.Integration) error
	DeleteIntegration(ctx context.Context, id sdktypes.IntegrationID) error
	GetIntegration(ctx context.Context, id sdktypes.IntegrationID) (sdktypes.Integration, error)
	ListIntegrations(ctx context.Context, filter sdkservices.ListIntegrationsFilter) ([]sdktypes.Integration, error)

	// -----------------------------------------------------------------------
	SetSecret(ctx context.Context, name string, data map[string]string) error
	GetSecret(ctx context.Context, name string) (map[string]string, error)
	AppendSecret(ctx context.Context, name, token string) error
	DeleteSecret(ctx context.Context, name string) error

	// -----------------------------------------------------------------------
	CreateSession(ctx context.Context, session sdktypes.Session) error
	GetSession(ctx context.Context, sessionID sdktypes.SessionID) (sdktypes.Session, error)
	GetSessionLog(ctx context.Context, sessionID sdktypes.SessionID) (sdktypes.SessionLog, error)
	UpdateSessionState(ctx context.Context, sessionID sdktypes.SessionID, state sdktypes.SessionState) error
	AddSessionPrint(ctx context.Context, sessionID sdktypes.SessionID, print string) error
	AddSessionStopRequested(ctx context.Context, sessionID sdktypes.SessionID, reason string) error
	ListSessions(ctx context.Context, f sdkservices.ListSessionsFilter) ([]sdktypes.Session, int, error)
	DeleteSession(ctx context.Context, sessionID sdktypes.SessionID) error

	// -----------------------------------------------------------------------
	CreateSessionCall(ctx context.Context, sessionID sdktypes.SessionID, data sdktypes.SessionCallSpec) error
	GetSessionCallSpec(ctx context.Context, sessionID sdktypes.SessionID, seq uint32) (sdktypes.SessionCallSpec, error)

	StartSessionCallAttempt(ctx context.Context, sessionID sdktypes.SessionID, seq uint32) (uint32, error)
	CompleteSessionCallAttempt(ctx context.Context, sessionID sdktypes.SessionID, seq, attempt uint32, complete sdktypes.SessionCallAttemptComplete) error
	GetSessionCallAttemptResult(ctx context.Context, sessionID sdktypes.SessionID, seq uint32, attempt int64 /* <0 for last */) (sdktypes.SessionCallAttemptResult, error)

	// -----------------------------------------------------------------------
	SaveSignal(ctx context.Context, signalID string, workflowID string, connectionID sdktypes.ConnectionID, eventName string) (string, error)
	GetSignal(ctx context.Context, signalID string) (scheme.Signal, error)
	RemoveSignal(ctx context.Context, signalID string) error
	ListSignalsWaitingOnConnection(ctx context.Context, connectionID sdktypes.ConnectionID, eventType string) ([]scheme.Signal, error)
}
