package dbgorm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (db *gormdb) getSession(ctx context.Context, sessionID string) (*scheme.Session, error) {
	return getOne(db.db, ctx, scheme.Session{}, "session_id = ?", sessionID)
}

func (db *gormdb) GetSession(ctx context.Context, id sdktypes.SessionID) (sdktypes.Session, error) {
	s, err := db.getSession(ctx, id.String())
	if s == nil || err != nil {
		return sdktypes.InvalidSession, translateError(err)
	}
	return scheme.ParseSession(*s)
}

func (db *gormdb) deleteSession(ctx context.Context, sessionID string) error {
	return delete(db.db, ctx, scheme.Session{}, "session_id = ?", sessionID)
}

func (db *gormdb) DeleteSession(ctx context.Context, sessionID sdktypes.SessionID) error {
	return translateError(db.deleteSession(ctx, sessionID.String()))
}

func (db *gormdb) GetSessionLog(ctx context.Context, sessionID sdktypes.SessionID) (sdktypes.SessionLog, error) {
	var rs []scheme.SessionLogRecord

	if err := db.db.WithContext(ctx).Where("session_id = ?", sessionID.String()).Find(&rs).Error; err != nil {
		return sdktypes.InvalidSessionLog, translateError(err)
	}

	prs, err := kittehs.TransformError(rs, scheme.ParseSessionLogRecord)

	return sdktypes.NewSessionLog(prs), err
}

func (db *gormdb) createSession(ctx context.Context, session *scheme.Session) error {
	return db.transaction(ctx, func(tx *tx) error {
		if err := tx.db.Create(session).Error; err != nil {
			return err
		}
		return addSessionLogRecord(
			tx.db,
			session.SessionID,
			sdktypes.NewStateSessionLogRecord(sdktypes.NewSessionStateCreated()),
		)
	})
}

func (db *gormdb) CreateSession(ctx context.Context, session sdktypes.Session) error {
	now := time.Now()

	s := scheme.Session{
		SessionID:        session.ID().String(),
		BuildID:          session.BuildID().String(),
		EnvID:            session.EnvID().String(),
		DeploymentID:     session.DeploymentID().String(),
		EventID:          session.EventID().String(),
		Entrypoint:       session.EntryPoint().CanonicalString(),
		CurrentStateType: int(sdktypes.SessionStateTypeCreated.ToProto()),
		Inputs:           kittehs.Must1(json.Marshal(session.Inputs())),
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	return translateError(db.createSession(ctx, &s))
}

func (db *gormdb) UpdateSessionState(ctx context.Context, sessionID sdktypes.SessionID, state sdktypes.SessionState) error {
	return translateError(db.transaction(ctx, func(tx *tx) error {
		r := scheme.Session{
			CurrentStateType: int(state.Type().ToProto()),
			UpdatedAt:        time.Now(),
		}

		sid := sessionID.String()
		if res := tx.db.Model(&r).Where("session_id = ?", sid).Updates(r); res.Error != nil {
			return res.Error
		} else if res.RowsAffected == 0 {
			return sdkerrors.ErrNotFound
		}

		return addSessionLogRecord(tx.db, sid, sdktypes.NewStateSessionLogRecord(state))
	}))
}

func addSessionLogRecord(tx *gorm.DB, sessionID string, logr sdktypes.SessionLogRecord) error {
	jsonData, err := json.Marshal(logr)
	if err != nil {
		return fmt.Errorf("marshal session log record: %w", err)
	}

	r := scheme.SessionLogRecord{SessionID: sessionID, Data: jsonData}
	return tx.Create(&r).Error
}

func (db *gormdb) AddSessionPrint(ctx context.Context, sessionID sdktypes.SessionID, print string) error {
	return translateError(
		addSessionLogRecord(db.db, sessionID.String(), sdktypes.NewPrintSessionLogRecord(print)),
	)
}

func (db *gormdb) AddSessionStopRequested(ctx context.Context, sessionID sdktypes.SessionID, reason string) error {
	return translateError(
		addSessionLogRecord(db.db, sessionID.String(), sdktypes.NewStopRequestedSessionLogRecord(reason)),
	)
}

func (db *gormdb) listSessions(ctx context.Context, f sdkservices.ListSessionsFilter) ([]scheme.Session, int, error) {
	var rs []scheme.Session

	q := db.db.WithContext(ctx)

	if f.DeploymentID.IsValid() {
		q = q.Where("deployment_id = ?", f.DeploymentID.String())
	}

	if f.EventID.IsValid() {
		q = q.Where("event_id = ?", f.EventID.String())
	}

	if f.BuildID.IsValid() {
		q = q.Where("build_id = ?", f.BuildID.String())
	}

	if f.StateType != sdktypes.SessionStateTypeUnspecified {
		q = q.Where("current_state_type = ?", f.StateType.ToProto())
	}

	if f.CountOnly {
		var n int64
		err := q.Model(&scheme.Session{}).Count(&n).Error
		return nil, int(n), err
	}

	if err := q.Order("created_at desc").Find(&rs).Error; err != nil {
		return nil, 0, err
	}

	// REVIEW: will the count be right in case of pagination?
	return rs, len(rs), nil
}

func (db *gormdb) ListSessions(ctx context.Context, f sdkservices.ListSessionsFilter) ([]sdktypes.Session, int, error) {
	rs, cnt, err := db.listSessions(ctx, f)
	if rs == nil { // no sessions to process. either error or count request
		return nil, cnt, translateError(err)
	}
	sessions, err := kittehs.TransformError(rs, scheme.ParseSession)
	return sessions, len(sessions), err
}

func (db *gormdb) CreateSessionCall(ctx context.Context, sessionID sdktypes.SessionID, spec sdktypes.SessionCallSpec) error {
	jsonSpec, err := json.Marshal(spec)
	if err != nil {
		return fmt.Errorf("marshal session call: %w", err)
	}

	return translateError(db.transaction(ctx, func(tx *tx) error {
		r := scheme.SessionCallSpec{
			SessionID: sessionID.String(),
			Seq:       spec.Seq(),
			Data:      jsonSpec,
		}

		if err := tx.db.Create(&r).Error; err != nil {
			return err
		}

		return addSessionLogRecord(tx.db, sessionID.String(), sdktypes.NewCallSpecSessionLogRecord(spec))
	}))
}

func (db *gormdb) GetSessionCallSpec(ctx context.Context, sessionID sdktypes.SessionID, seq uint32) (sdktypes.SessionCallSpec, error) {
	var r scheme.SessionCallSpec
	if err := db.db.
		Where("session_id = ?", sessionID.String()).
		Where("seq = ?", seq).
		First(&r).
		Error; err != nil {
		return sdktypes.InvalidSessionCallSpec, translateError(err)
	}

	spec, err := scheme.ParseSessionCallSpec(r)
	if err != nil {
		return sdktypes.InvalidSessionCallSpec, err
	}

	return spec, nil
}

func (db *gormdb) StartSessionCallAttempt(ctx context.Context, sessionID sdktypes.SessionID, seq uint32) (attempt uint32, err error) {
	err = translateError(db.transaction(ctx, func(tx *tx) error {
		var err error
		if attempt, err = countCallAttemps(tx.db, sessionID, seq); err != nil {
			return err
		}

		obj := kittehs.Must1(sdktypes.SessionCallAttemptStartFromProto(&sdktypes.SessionCallAttemptStartPB{
			StartedAt: timestamppb.Now(),
			Num:       attempt,
		}))

		json, err := json.Marshal(obj)
		if err != nil {
			return err
		}

		sid := sessionID.String()
		r := scheme.SessionCallAttempt{
			SessionID: sid,
			Seq:       seq,
			Attempt:   attempt,
			Start:     json,
		}

		if err := tx.db.Create(&r).Error; err != nil {
			return err
		}

		return addSessionLogRecord(tx.db, sessionID.String(), sdktypes.NewCallAttemptStartSessionLogRecord(obj))
	}))

	return
}

func (db *gormdb) CompleteSessionCallAttempt(ctx context.Context, sessionID sdktypes.SessionID, seq, attempt uint32, complete sdktypes.SessionCallAttemptComplete) error {
	return translateError(db.transaction(ctx, func(tx *tx) error {
		json, err := json.Marshal(complete)
		if err != nil {
			return fmt.Errorf("marshal session call attempt complete: %w", err)
		}

		r := scheme.SessionCallAttempt{
			Complete: json,
		}

		if res := tx.db.Model(&r).Where("session_id = ? AND seq = ? AND attempt = ?", sessionID.String(), seq, attempt).Updates(r); res.Error != nil {
			return res.Error
		} else if res.RowsAffected == 0 {
			return sdkerrors.ErrNotFound
		}

		return addSessionLogRecord(tx.db, sessionID.String(), sdktypes.NewCallAttemptCompleteSessionLogRecord(complete))
	}))
}

// attempt = -1: latest.
// attempt >= 0: specific attempt.
func (db *gormdb) GetSessionCallAttemptResult(ctx context.Context, sessionID sdktypes.SessionID, seq uint32, attempt int64) (sdktypes.SessionCallAttemptResult, error) {
	q := db.db.Where("session_id = ? AND seq = ?", sessionID.String(), seq)

	if attempt == -1 {
		q = q.Order(clause.OrderByColumn{Column: clause.Column{Name: "attempt"}, Desc: true})
	} else if attempt >= 0 {
		q = q.Where("attempt = ?", attempt)
	} else {
		return sdktypes.InvalidSessionCallAttemptResult, sdkerrors.NewInvalidArgumentError("attempt must be either -1 or >= 0, got %d", attempt)
	}

	var r scheme.SessionCallAttempt
	if err := q.First(&r).Error; err != nil {
		return sdktypes.InvalidSessionCallAttemptResult, translateError(err)
	}

	if r.Complete == nil {
		return sdktypes.InvalidSessionCallAttemptResult, sdkerrors.ErrNotFound
	}

	complete, err := scheme.ParseSessionCallAttemptComplete(r)
	if err != nil {
		return sdktypes.InvalidSessionCallAttemptResult, err
	}

	return complete.Result(), nil
}

func countCallAttemps(db *gorm.DB, sessionID sdktypes.SessionID, seq uint32) (uint32, error) {
	var n int64

	if err := db.
		Model(&scheme.SessionCallAttempt{}).
		Where("session_id = ? AND seq = ?", sessionID.String(), seq).
		Count(&n).
		Error; err != nil {
		return 0, err
	}

	return uint32(n), nil
}
