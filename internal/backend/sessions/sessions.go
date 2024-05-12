package sessions

import (
	"context"
	"fmt"

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
	config *Config
	z      *zap.Logger
	svcs   *sessionsvcs.Svcs

	workflows sessionworkflows.Workflows
	calls     sessioncalls.Calls
}

var _ Sessions = (*sessions)(nil)

func New(z *zap.Logger, config *Config, db db.DB, svcs sessionsvcs.Svcs) Sessions {
	return &sessions{
		config: config,
		svcs:   &svcs,
		z:      z,
	}
}

func (s *sessions) StartWorkers(ctx context.Context) error {
	s.calls = sessioncalls.New(s.z.Named("sessionworkflows"), s.config.Calls, s.svcs)
	s.workflows = sessionworkflows.New(s.z.Named("sessionworkflows"), s.config.Workflows, s, s.svcs, s.calls)

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

func (s *sessions) List(ctx context.Context, filter sdkservices.ListSessionsFilter) (sdkservices.ListSessionResult, error) {
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
		return fmt.Errorf("%w: cannot delete session, invalid state: %s, session_id: %s", sdkerrors.ErrFailedPrecondition, session.State(), sessionID.String())
	}

	err = s.svcs.DB.DeleteSession(ctx, sessionID)
	s.z.With(zap.String("session_id", sessionID.String())).Info("delete")
	return err
}

func (s *sessions) Start(ctx context.Context, session sdktypes.Session) (sdktypes.SessionID, error) {
	if session.ID().IsValid() {
		return sdktypes.InvalidSessionID, sdkerrors.NewInvalidArgumentError("session id is not nil")
	}

	session = session.WithNewID()

	if err := s.svcs.DB.CreateSession(ctx, session); err != nil {
		return sdktypes.InvalidSessionID, fmt.Errorf("db.create_session: %w", err)
	}

	if err := s.workflows.StartWorkflow(ctx, session, s.config.Debug); err != nil {
		if uerr := s.svcs.DB.UpdateSessionState(
			ctx,
			session.ID(),
			sdktypes.NewSessionStateError(fmt.Errorf("execute workflow: %w", err), nil),
		); uerr != nil {
			s.z.With(zap.String("session_id", session.ID().String())).Error("update session", zap.Error(err))
		}

		return sdktypes.InvalidSessionID, fmt.Errorf("start workflow: %w", err)
	}

	return session.ID(), nil
}
