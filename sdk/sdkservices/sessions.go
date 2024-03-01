package sdkservices

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type ListSessionsFilter struct {
	DeploymentID sdktypes.DeploymentID
	EnvID        sdktypes.EnvID
	EventID      sdktypes.EventID
	StateType    sdktypes.SessionStateType
	CountOnly    bool
}

type Sessions interface {
	Start(ctx context.Context, session sdktypes.Session) (sdktypes.SessionID, error)
	Stop(ctx context.Context, sessionID sdktypes.SessionID, reason string, force bool) error
	List(ctx context.Context, filter ListSessionsFilter) ([]sdktypes.Session, int, error)
	Get(ctx context.Context, sessionID sdktypes.SessionID) (sdktypes.Session, error)
	GetLog(ctx context.Context, sessionID sdktypes.SessionID) (sdktypes.SessionLog, error)
	Delete(ctx context.Context, sessionID sdktypes.SessionID) error
}
