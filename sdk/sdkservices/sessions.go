package sdkservices

import (
	"context"
	"time"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type ListSessionsFilter struct {
	DeploymentID sdktypes.DeploymentID
	OrgID        sdktypes.OrgID
	ProjectID    sdktypes.ProjectID
	EventID      sdktypes.EventID
	BuildID      sdktypes.BuildID
	StateType    sdktypes.SessionStateType
	CountOnly    bool

	sdktypes.PaginationRequest
}

func (f ListSessionsFilter) AnyIDSpecified() bool {
	return f.DeploymentID.IsValid() || f.OrgID.IsValid() || f.ProjectID.IsValid() || f.EventID.IsValid() || f.BuildID.IsValid()
}

type ListSessionResult struct {
	Sessions []sdktypes.Session
	sdktypes.PaginationResult
}

type SessionLogRecordsFilter struct {
	SessionID sdktypes.SessionID
	Types     sdktypes.SessionLogRecordType // bitmask
	sdktypes.PaginationRequest
}

type GetLogResults struct {
	Records []sdktypes.SessionLogRecord
	sdktypes.PaginationResult
}

type SessionPrint struct {
	Timestamp time.Time
	Value     sdktypes.Value
}

type GetPrintsResults struct {
	Prints []*SessionPrint
	sdktypes.PaginationResult
}

type Sessions interface {
	Start(ctx context.Context, session sdktypes.Session) (sdktypes.SessionID, error)
	// Will always try first to gracefully terminate the session.
	// Blocks only if `force` and forceDelay > 0`.
	Stop(ctx context.Context, sessionID sdktypes.SessionID, reason string, force bool, forceDelay time.Duration) error
	// List returns sessions without their data.
	List(ctx context.Context, filter ListSessionsFilter) (*ListSessionResult, error)
	Get(ctx context.Context, sessionID sdktypes.SessionID) (sdktypes.Session, error)
	GetLog(ctx context.Context, filter SessionLogRecordsFilter) (*GetLogResults, error)
	DownloadLogs(ctx context.Context, sessionID sdktypes.SessionID) ([]byte, error)
	GetPrints(ctx context.Context, sid sdktypes.SessionID, pagination sdktypes.PaginationRequest) (*GetPrintsResults, error)
	Delete(ctx context.Context, sessionID sdktypes.SessionID) error
}
