package sessions

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	akCtx "go.autokitteh.dev/autokitteh/internal/backend/context"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessioncalls"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionsvcs"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionworkflows"
	"go.autokitteh.dev/autokitteh/internal/backend/telemetry"
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
	l      *zap.Logger
	svcs   *sessionsvcs.Svcs

	workflows sessionworkflows.Workflows
	calls     sessioncalls.Calls
	telemetry *telemetry.Telemetry
}

var _ Sessions = (*sessions)(nil)

func New(l *zap.Logger, config *Config, db db.DB, svcs sessionsvcs.Svcs, telemetry *telemetry.Telemetry) Sessions {
	return &sessions{
		config:    config,
		svcs:      &svcs,
		l:         l,
		telemetry: telemetry,
	}
}

func (s *sessions) StartWorkers(ctx context.Context) error {
	s.calls = sessioncalls.New(s.l.Named("sessionworkflows"), s.config.Calls, s.svcs)
	s.workflows = sessionworkflows.New(s.l.Named("sessionworkflows"), s.config.Workflows, s, s.svcs, s.calls, s.telemetry)

	if !s.config.EnableWorker {
		s.l.Info("Session worker: disabled")
		return nil
	}

	s.l.Info("Session worker: enabled")

	if err := s.workflows.StartWorkers(ctx); err != nil {
		return fmt.Errorf("workflow workflows: %w", err)
	}

	if err := s.calls.StartWorkers(ctx); err != nil {
		return fmt.Errorf("activity workflows: %w", err)
	}

	return nil
}

func (s *sessions) GetLog(ctx context.Context, filter sdkservices.ListSessionLogRecordsFilter) (sdkservices.GetLogResults, error) {
	return s.svcs.DB.GetSessionLog(ctx, filter)
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

	// delete only finalized sessions.
	if state := session.State(); !state.IsFinal() {
		return fmt.Errorf("%w: cannot delete session while in progress: %s, session_id: %v", sdkerrors.ErrFailedPrecondition, state, sessionID)
	}

	if err = s.svcs.DB.DeleteSession(ctx, sessionID); err != nil {
		return err
	}

	s.l.Sugar().With("session_id", sessionID).Infof("deleted %v", sessionID)

	return nil
}

func (s *sessions) Start(ctx context.Context, session sdktypes.Session) (sdktypes.SessionID, error) {
	if session.ID().IsValid() {
		return sdktypes.InvalidSessionID, sdkerrors.NewInvalidArgumentError("session id is not nil")
	}

	if err := session.Strict(); err != nil {
		return sdktypes.InvalidSessionID, err
	}

	session = session.WithNewID()
	sid := session.ID()
	l := s.l.With(zap.Any("session_id", sid))

	ctx = akCtx.WithRequestOrginator(ctx, akCtx.SessionWorkflow)

	if pid := session.ProjectID(); pid.IsValid() {
		var err error
		if ctx, err = akCtx.WithOwnershipOf(ctx, s.svcs.DB.GetOwnership, pid.UUIDValue()); err != nil {
			return sdktypes.InvalidSessionID, fmt.Errorf("ownership: %w", err)
		}
	}

	if err := s.svcs.DB.CreateSession(ctx, session); err != nil {
		return sdktypes.InvalidSessionID, fmt.Errorf("start session: %w", err)
	}

	if err := s.workflows.StartWorkflow(ctx, session); err != nil {
		err = fmt.Errorf("start workflow: %w", err)
		if uerr := s.svcs.DB.UpdateSessionState(ctx, session.ID(), sdktypes.NewSessionStateError(err, nil)); uerr != nil {
			l.Sugar().With("err", err).Error("update session state: %v")
		}
		return sdktypes.InvalidSessionID, fmt.Errorf("start workflow: %w", err)
	}

	return session.ID(), nil
}
