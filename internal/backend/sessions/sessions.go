package sessions

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/client"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessioncalls"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionsvcs"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionworkflows"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Sessions interface {
	sdkservices.Sessions

	StartWorkers(context.Context) error
}
type sessions struct {
	config   *Config
	temporal client.Client
	z        *zap.Logger
	svcs     *sessionsvcs.Svcs

	workflows sessionworkflows.Workflows
	calls     sessioncalls.Calls
}

var _ Sessions = (*sessions)(nil)

func New(z *zap.Logger, config *Config, temporal client.Client, db db.DB, svcs sessionsvcs.Svcs) Sessions {
	sessions := &sessions{
		config:   config,
		temporal: temporal,
		svcs:     &svcs,
		z:        z,
	}

	sessions.calls = sessioncalls.New(z.Named("sessionworkflows"), config.Calls, &svcs)
	sessions.workflows = sessionworkflows.New(z.Named("sessionworkflows"), config.Workflows, sessions, &svcs, sessions.calls)

	return sessions
}

func (s *sessions) StartWorkers(ctx context.Context) error {
	if err := s.workflows.StartWorkers(ctx); err != nil {
		return fmt.Errorf("workflow workflows: %w", err)
	}

	if err := s.calls.StartWorkers(ctx); err != nil {
		return fmt.Errorf("activity workflows: %w", err)
	}

	return nil
}

func (s *sessions) GetLog(ctx context.Context, sessionID sdktypes.SessionID) (sdktypes.SessionLog, error) {
	return s.svcs.DB.GetSessionLog(ctx, sessionID)
}

func (s *sessions) Get(ctx context.Context, sessionID sdktypes.SessionID) (sdktypes.Session, error) {
	return s.svcs.DB.GetSession(ctx, sessionID)
}

func (s *sessions) Stop(ctx context.Context, sessionID sdktypes.SessionID, reason string, force bool) error {
	return s.workflows.StopWorkflow(ctx, sessionID, reason, force)
}

func (s *sessions) List(ctx context.Context, filter sdkservices.ListSessionsFilter) ([]sdktypes.Session, int, error) {
	return s.svcs.DB.ListSessions(ctx, filter)
}

func (s *sessions) Delete(ctx context.Context, sessionID sdktypes.SessionID) error {
	session, err := s.Get(ctx, sessionID)
	if err != nil {
		return err
	}

	// delete only failed or finished sessions
	state := session.State()
	if state != sdktypes.SessionStateTypeCompleted && state != sdktypes.SessionStateTypeError {
		return fmt.Errorf("%w: cannot delete active session: session_id: %s", sdkerrors.ErrFailedPrecondition, sessionID.String())
	}

	err = s.svcs.DB.DeleteSession(ctx, sessionID)
	s.z.With(zap.String("session_id", sessionID.String())).Info("delete")
	return err
}

func (s *sessions) Start(ctx context.Context, session sdktypes.Session) (sdktypes.SessionID, error) {
	sessionID := session.ID()
	if sessionID.IsValid() {
		return sdktypes.InvalidSessionID, sdkerrors.NewInvalidArgumentError("session id is not nil")
	}

	session = session.WithNewID()

	if err := s.svcs.DB.CreateSession(ctx, session); err != nil {
		return sdktypes.InvalidSessionID, fmt.Errorf("db.create_session: %w", err)
	}

	if err := s.workflows.StartWorkflow(ctx, session, s.config.Debug); err != nil {
		if uerr := s.svcs.DB.UpdateSessionState(
			ctx,
			sessionID,
			sdktypes.NewSessionStateError(fmt.Errorf("execute workflow: %w", err), nil),
		); uerr != nil {
			s.z.With(zap.String("session_id", sessionID.String())).Error("update session", zap.Error(err))
		}

		return sdktypes.InvalidSessionID, fmt.Errorf("start workflow: %w", err)
	}

	return sessionID, nil
}
