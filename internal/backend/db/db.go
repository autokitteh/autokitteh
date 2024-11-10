package db

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/health/healthreporter"
	"go.autokitteh.dev/autokitteh/internal/backend/types"
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
	MigrationRequired(context.Context) (bool, int64, error)
	Migrate(context.Context) error
	Debug() DB

	healthreporter.HealthReporter

	GormDB() *gorm.DB

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
	SetVars(context.Context, []sdktypes.Var) error
	GetVars(context.Context, sdktypes.VarScopeID, []sdktypes.Symbol) ([]sdktypes.Var, error)
	CountVars(context.Context, sdktypes.VarScopeID) (int, error)
	DeleteVars(context.Context, sdktypes.VarScopeID, []sdktypes.Symbol) error
	FindConnectionIDsByVar(context.Context, sdktypes.IntegrationID, sdktypes.Symbol, string) ([]sdktypes.ConnectionID, error)

	// -----------------------------------------------------------------------
	// This is idempotent.
	SaveEvent(context.Context, sdktypes.Event) error
	GetEventByID(context.Context, sdktypes.EventID) (sdktypes.Event, error)
	ListEvents(context.Context, sdkservices.ListEventsFilter) ([]sdktypes.Event, error)
	GetLatestEventSequence(context.Context) (uint64, error)

	// -----------------------------------------------------------------------
	CreateTrigger(context.Context, sdktypes.Trigger) error
	UpdateTrigger(context.Context, sdktypes.Trigger) error
	GetTriggerByID(context.Context, sdktypes.TriggerID) (sdktypes.Trigger, error)
	DeleteTrigger(context.Context, sdktypes.TriggerID) error
	ListTriggers(context.Context, sdkservices.ListTriggersFilter) ([]sdktypes.Trigger, error)
	GetTriggerByWebhookSlug(ctx context.Context, slug string) (sdktypes.Trigger, error)

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
	GetConnections(ctx context.Context, ids []sdktypes.ConnectionID) ([]sdktypes.Connection, error)
	ListConnections(ctx context.Context, filter sdkservices.ListConnectionsFilter, idsOnly bool) ([]sdktypes.Connection, error)

	// -----------------------------------------------------------------------
	GetDeployment(ctx context.Context, id sdktypes.DeploymentID) (sdktypes.Deployment, error)

	// Returns deployments in ascending order by creation time.
	ListDeployments(ctx context.Context, filter sdkservices.ListDeploymentsFilter) ([]sdktypes.Deployment, error)
	UpdateDeploymentState(ctx context.Context, id sdktypes.DeploymentID, state sdktypes.DeploymentState) (oldState sdktypes.DeploymentState, err error)
	CreateDeployment(ctx context.Context, deployment sdktypes.Deployment) error
	DeleteDeployment(ctx context.Context, deploymentID sdktypes.DeploymentID) error
	DeploymentHasActiveSessions(ctx context.Context, deploymentID sdktypes.DeploymentID) (bool, error)

	// -----------------------------------------------------------------------
	CreateSession(ctx context.Context, session sdktypes.Session) error
	GetSession(ctx context.Context, sessionID sdktypes.SessionID) (sdktypes.Session, error)
	GetSessionLog(ctx context.Context, filter sdkservices.ListSessionLogRecordsFilter) (sdkservices.GetLogResults, error)
	UpdateSessionState(ctx context.Context, sessionID sdktypes.SessionID, state sdktypes.SessionState) error
	AddSessionPrint(ctx context.Context, sessionID sdktypes.SessionID, print string) error
	AddSessionStopRequest(ctx context.Context, sessionID sdktypes.SessionID, reason string) error
	ListSessions(ctx context.Context, f sdkservices.ListSessionsFilter) (sdkservices.ListSessionResult, error)
	DeleteSession(ctx context.Context, sessionID sdktypes.SessionID) error

	// -----------------------------------------------------------------------
	CreateSessionCall(ctx context.Context, sessionID sdktypes.SessionID, data sdktypes.SessionCallSpec) error
	GetSessionCallSpec(ctx context.Context, sessionID sdktypes.SessionID, seq uint32) (sdktypes.SessionCallSpec, error)

	StartSessionCallAttempt(ctx context.Context, sessionID sdktypes.SessionID, seq uint32) (uint32, error)
	CompleteSessionCallAttempt(ctx context.Context, sessionID sdktypes.SessionID, seq, attempt uint32, complete sdktypes.SessionCallAttemptComplete) error
	GetSessionCallAttemptResult(ctx context.Context, sessionID sdktypes.SessionID, seq uint32, attempt int64 /* <0 for last */) (sdktypes.SessionCallAttemptResult, error)

	// -----------------------------------------------------------------------
	// TODO(ENG-917): Do not expose scheme outside of DB.
	SaveSignal(ctx context.Context, signal *types.Signal) error
	GetSignal(ctx context.Context, signalID uuid.UUID) (*types.Signal, error)
	RemoveSignal(ctx context.Context, signalID uuid.UUID) error
	ListWaitingSignals(ctx context.Context, dstID sdktypes.EventDestinationID) ([]*types.Signal, error)

	// -----------------------------------------------------------------------
	CreateUser(ctx context.Context, user sdktypes.User) (sdktypes.UserID, error)
	GetUserByEmail(ctx context.Context, email string) (sdktypes.User, error)
	GetUserByID(ctx context.Context, id sdktypes.UserID) (sdktypes.User, error)

	// -----------------------------------------------------------------------
	SetSecret(ctx context.Context, key string, value string) error
	GetSecret(ctx context.Context, key string) (string, error)
	DeleteSecret(ctx context.Context, key string) error

	// -----------------------------------------------------------------------
	GetOwnership(ctx context.Context, entityID sdktypes.UUID) (sdktypes.User, error)

	SetValue(ctx context.Context, pid sdktypes.ProjectID, key string, v sdktypes.Value) error
	GetValue(ctx context.Context, pid sdktypes.ProjectID, key string) (sdktypes.Value, error)
	ListValues(ctx context.Context, pid sdktypes.ProjectID) (map[string]sdktypes.Value, error)
}
