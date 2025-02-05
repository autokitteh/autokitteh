package sessions

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authz"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessioncalls"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionsvcs"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionworkflows"
	"go.autokitteh.dev/autokitteh/internal/backend/telemetry"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
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

func (s *sessions) GetPrints(ctx context.Context, sid sdktypes.SessionID, pagination sdktypes.PaginationRequest) (*sdkservices.GetPrintsResults, error) {
	if err := authz.CheckContext(ctx, sid, "read:get-prints", authz.WithData("pagination", pagination), authz.WithConvertForbiddenToNotFound); err != nil {
		return nil, err
	}

	lr, err := s.svcs.DB.GetSessionLog(ctx, sdkservices.SessionLogRecordsFilter{
		SessionID:         sid,
		Types:             sdktypes.PrintSessionLogRecordType,
		PaginationRequest: pagination,
	})
	if err != nil {
		return nil, err
	}

	prints := kittehs.Transform(lr.Records, func(r sdktypes.SessionLogRecord) *sdkservices.SessionPrint {
		p, _ := r.GetPrint()

		return &sdkservices.SessionPrint{
			Timestamp: r.Timestamp(),
			Value:     sdktypes.NewStringValue(p),
		}
	})

	return &sdkservices.GetPrintsResults{
		Prints:           prints,
		PaginationResult: lr.PaginationResult,
	}, nil
}

func (s *sessions) GetLog(ctx context.Context, filter sdkservices.SessionLogRecordsFilter) (*sdkservices.GetLogResults, error) {
	if err := authz.CheckContext(ctx, filter.SessionID, "read:get-log", authz.WithData("filter", filter), authz.WithConvertForbiddenToNotFound); err != nil {
		return nil, err
	}

	return s.svcs.DB.GetSessionLog(ctx, filter)
}

func (s *sessions) Get(ctx context.Context, sessionID sdktypes.SessionID) (sdktypes.Session, error) {
	if err := authz.CheckContext(ctx, sessionID, "read:get", authz.WithConvertForbiddenToNotFound); err != nil {
		return sdktypes.InvalidSession, err
	}

	return s.svcs.DB.GetSession(ctx, sessionID)
}

func (s *sessions) Stop(ctx context.Context, sessionID sdktypes.SessionID, reason string, force bool, cancelTimeout time.Duration) error {
	if err := authz.CheckContext(
		ctx,
		sessionID,
		"write:stop",
		authz.WithData("force", force),
		authz.WithData("cancel_timeout", cancelTimeout),
	); err != nil {
		return err
	}

	return s.workflows.StopWorkflow(ctx, sessionID, reason, force, cancelTimeout)
}

func (s *sessions) List(ctx context.Context, filter sdkservices.ListSessionsFilter) (*sdkservices.ListSessionResult, error) {
	if !filter.AnyIDSpecified() {
		filter.OrgID = authcontext.GetAuthnInferredOrgID(ctx)
	}

	if err := authz.CheckContext(
		ctx,
		sdktypes.InvalidSessionID,
		"list",
		authz.WithData("filter", filter),
		authz.WithAssociationWithID("deployment", filter.DeploymentID),
		authz.WithAssociationWithID("project", filter.ProjectID),
		authz.WithAssociationWithID("org", filter.OrgID),
		authz.WithAssociationWithID("event", filter.EventID),
		authz.WithAssociationWithID("build", filter.BuildID),
	); err != nil {
		return nil, err
	}

	return s.svcs.DB.ListSessions(ctx, filter)
}

func (s *sessions) Delete(ctx context.Context, sessionID sdktypes.SessionID) error {
	if err := authz.CheckContext(ctx, sessionID, "delete:delete"); err != nil {
		return err
	}

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
	if err := authz.CheckContext(
		ctx,
		sdktypes.InvalidSessionID,
		"create:start",
		authz.WithData("session", session),
		authz.WithAssociationWithID("build", session.BuildID()),
		authz.WithAssociationWithID("deployment", session.DeploymentID()),
		authz.WithAssociationWithID("event", session.EventID()),
		authz.WithAssociationWithID("parent_session", session.ParentSessionID()),
		authz.WithAssociationWithID("project", session.ProjectID()),
	); err != nil {
		return sdktypes.InvalidSessionID, err
	}

	if session.ID().IsValid() {
		return sdktypes.InvalidSessionID, sdkerrors.NewInvalidArgumentError("session id is not nil")
	}

	if err := session.Strict(); err != nil {
		return sdktypes.InvalidSessionID, err
	}

	session = session.WithNewID()
	sid := session.ID()
	l := s.l.With(zap.Any("session_id", sid))

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
