package sdkservices

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type ListSessionsFilter struct {
	DeploymentID sdktypes.DeploymentID
	EnvID        sdktypes.EnvID
	EventID      sdktypes.EventID
	BuildID      sdktypes.BuildID
	StateType    sdktypes.SessionStateType
	CountOnly    bool

	sdktypes.PaginationRequest
}

type ListSessionResult struct {
	Sessions []sdktypes.Session
	sdktypes.PaginationResult
}

type ListSessionLogRecordsFilter struct {
	SessionID    sdktypes.SessionID
	IgnorePrints bool
	sdktypes.PaginationRequest
}

type GetLogResults struct {
	Log sdktypes.SessionLog
	sdktypes.PaginationResult
}

type Sessions interface {
	Start(ctx context.Context, session sdktypes.Session) (sdktypes.SessionID, error)
	Stop(ctx context.Context, sessionID sdktypes.SessionID, reason string, force bool) error
	// List returns sessions without their data.
	List(ctx context.Context, filter ListSessionsFilter) (ListSessionResult, error)
	Get(ctx context.Context, sessionID sdktypes.SessionID) (sdktypes.Session, error)
	GetLog(ctx context.Context, filter ListSessionLogRecordsFilter) (GetLogResults, error)
	Delete(ctx context.Context, sessionID sdktypes.SessionID) error
}
