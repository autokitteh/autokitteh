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
	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (db *gormdb) getSession(ctx context.Context, sessionID sdktypes.UUID) (*scheme.SessionWithInputs, error) {
	return getOne[scheme.SessionWithInputs](db.db, ctx, "session_id = ?", sessionID)
}

func (db *gormdb) GetSession(ctx context.Context, id sdktypes.SessionID) (sdktypes.Session, error) {
	s, err := db.getSession(ctx, id.UUIDValue())
	if s == nil || err != nil {
		return sdktypes.InvalidSession, translateError(err)
	}
	return scheme.ParseSessionWithInputs(*s)
}

func (db *gormdb) deleteSession(ctx context.Context, sessionID sdktypes.UUID) error {
	return delete[scheme.Session](db.db, ctx, "session_id = ?", sessionID)
}

func (db *gormdb) DeleteSession(ctx context.Context, sessionID sdktypes.SessionID) error {
	return translateError(db.deleteSession(ctx, sessionID.UUIDValue()))
}

func (db *gormdb) GetSessionLog(ctx context.Context, sessionID sdktypes.SessionID) (sdktypes.SessionLog, error) {
	var rs []scheme.SessionLogRecord

	if err := db.db.WithContext(ctx).Where("session_id = ?", sessionID.UUIDValue()).Find(&rs).Error; err != nil {
		return sdktypes.InvalidSessionLog, translateError(err)
	}

	prs, err := kittehs.TransformError(rs, scheme.ParseSessionLogRecord)

	return sdktypes.NewSessionLog(prs), err
}

func (db *gormdb) createSession(ctx context.Context, session *scheme.SessionWithInputs) error {
	return db.transaction(ctx, func(tx *tx) error {
		if err := tx.db.Create(session).Error; err != nil {
			return err
		}
		return addSessionLogRecord(
			tx.db,
			session.SessionID,
			sdktypes.NewStateSessionLogRecord(sdktypes.NewSessionStateCreated()).WithProcessID(fixtures.ProcessID()),
		)
	})
}

func (db *gormdb) CreateSession(ctx context.Context, session sdktypes.Session) error {
	if err := session.Strict(); err != nil {
		return err
	}

	now := time.Now()

	cinputs, err := scheme.CompressJSON(session.Inputs())
	if err != nil {
		return fmt.Errorf("compress inputs: %w", err)
	}

	s := scheme.SessionWithInputs{
		Session: scheme.Session{
			SessionID:        session.ID().UUIDValue(),
			BuildID:          scheme.UUIDOrNil(session.BuildID().UUIDValue()),
			EnvID:            scheme.UUIDOrNil(session.EnvID().UUIDValue()),
			DeploymentID:     scheme.UUIDOrNil(session.DeploymentID().UUIDValue()),
			EventID:          scheme.UUIDOrNil(session.EventID().UUIDValue()),
			Entrypoint:       session.EntryPoint().CanonicalString(),
			CurrentStateType: int(sdktypes.SessionStateTypeCreated.ToProto()),
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		CompressedInputs: cinputs,
	}
	return translateError(db.createSession(ctx, &s))
}

func (db *gormdb) UpdateSessionState(ctx context.Context, sessionID sdktypes.SessionID, state sdktypes.SessionState) error {
	return translateError(db.transaction(ctx, func(tx *tx) error {
		r := scheme.Session{
			CurrentStateType: int(state.Type().ToProto()),
			UpdatedAt:        time.Now(),
		}

		sid := sessionID.UUIDValue()
		if res := tx.db.Model(&r).Where("session_id = ?", sid).Updates(r); res.Error != nil {
			return res.Error
		} else if res.RowsAffected == 0 {
			return sdkerrors.ErrNotFound
		}

		return addSessionLogRecord(tx.db, sid, sdktypes.NewStateSessionLogRecord(state).WithProcessID(fixtures.ProcessID()))
	}))
}

func addSessionLogRecordDB(tx *gorm.DB, logr *scheme.SessionLogRecord) error {
	return tx.Create(logr).Error
}

func addSessionLogRecord(tx *gorm.DB, sessionID sdktypes.UUID, logr sdktypes.SessionLogRecord) error {
	jsonData, err := json.Marshal(logr)
	if err != nil {
		return fmt.Errorf("marshal session log record: %w", err)
	}
	return addSessionLogRecordDB(tx, &scheme.SessionLogRecord{SessionID: sessionID, Data: jsonData})
}

func (db *gormdb) AddSessionPrint(ctx context.Context, sessionID sdktypes.SessionID, print string) error {
	return translateError(
		addSessionLogRecord(db.db, sessionID.UUIDValue(), sdktypes.NewPrintSessionLogRecord(print).WithProcessID(fixtures.ProcessID())),
	)
}

func (db *gormdb) AddSessionStopRequest(ctx context.Context, sessionID sdktypes.SessionID, reason string) error {
	return translateError(
		addSessionLogRecord(db.db, sessionID.UUIDValue(), sdktypes.NewStopRequestSessionLogRecord(reason).WithProcessID(fixtures.ProcessID())),
	)
}

func (db *gormdb) listSessions(ctx context.Context, f sdkservices.ListSessionsFilter) ([]scheme.Session, int, error) {
	var rs []scheme.Session

	q := db.db.WithContext(ctx)

	if f.DeploymentID.IsValid() {
		q = q.Where("deployment_id = ?", f.DeploymentID.UUIDValue())
	}

	if f.EventID.IsValid() {
		q = q.Where("event_id = ?", f.EventID.UUIDValue())
	}

	if f.BuildID.IsValid() {
		q = q.Where("build_id = ?", f.BuildID.UUIDValue())
	}

	if f.StateType != sdktypes.SessionStateTypeUnspecified {
		q = q.Where("current_state_type = ?", f.StateType.ToProto())
	}

	if f.CountOnly {
		var n int64
		err := q.Model(&scheme.Session{}).Count(&n).Error
		return nil, int(n), err
	}

	if f.PageSize != 0 {
		q = q.Limit(int(f.PageSize))
	}

	if f.Skip != 0 {
		q = q.Offset(int(f.Skip))
	}

	if f.PageToken != "" {
		q = q.Where("session_id < ?", f.PageToken)
	}

	// Double order in case we have two rows with the same created_at, then solve order by session_id
	if err := q.
		Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}, Desc: true}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "session_id"}, Desc: true}).
		Find(&rs).Error; err != nil {
		return nil, 0, err
	}

	// REVIEW: will the count be right in case of pagination?
	return rs, len(rs), nil
}

func (db *gormdb) ListSessions(ctx context.Context, f sdkservices.ListSessionsFilter) (sdkservices.ListSessionResult, error) {
	rs, cnt, err := db.listSessions(ctx, f)
	if rs == nil { // no sessions to process. either error or count request
		return sdkservices.ListSessionResult{}, translateError(err)
	}
	sessions, err := kittehs.TransformError(rs, scheme.ParseSession)

	// Only if we have a full page, there might be more sessions
	nextPageToken := ""
	if len(sessions) == int(f.PageSize) && len(sessions) > 0 {
		nextPageToken = sessions[len(sessions)-1].ID().UUIDValue().String()
	}

	res := sdkservices.ListSessionResult{
		Sessions:         sessions,
		PaginationResult: sdktypes.PaginationResult{TotalCount: cnt, NextPageToken: nextPageToken},
	}

	return res, err
}

func (db *gormdb) CreateSessionCall(ctx context.Context, sessionID sdktypes.SessionID, spec sdktypes.SessionCallSpec) error {
	cdata, err := scheme.CompressJSON(spec)
	if err != nil {
		return fmt.Errorf("marshal session call: %w", err)
	}

	return translateError(db.transaction(ctx, func(tx *tx) error {
		r := scheme.SessionCallSpec{
			SessionID:      sessionID.UUIDValue(),
			Seq:            spec.Seq(),
			CompressedData: cdata,
		}

		if err := tx.db.Create(&r).Error; err != nil {
			return err
		}

		return addSessionLogRecord(tx.db, sessionID.UUIDValue(), sdktypes.NewCallSpecSessionLogRecord(spec).WithProcessID(fixtures.ProcessID()))
	}))
}

func (db *gormdb) GetSessionCallSpec(ctx context.Context, sessionID sdktypes.SessionID, seq uint32) (sdktypes.SessionCallSpec, error) {
	var r scheme.SessionCallSpec
	if err := db.db.
		Where("session_id = ?", sessionID.UUIDValue()).
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

		sid := sessionID.UUIDValue()
		r := scheme.SessionCallAttempt{
			SessionID: sid,
			Seq:       seq,
			Attempt:   attempt,
			Start:     json,
		}

		if err := tx.db.Create(&r).Error; err != nil {
			return err
		}

		return addSessionLogRecord(tx.db, sid, sdktypes.NewCallAttemptStartSessionLogRecord(obj).WithProcessID(fixtures.ProcessID()))
	}))

	return
}

func (db *gormdb) CompleteSessionCallAttempt(ctx context.Context, sessionID sdktypes.SessionID, seq, attempt uint32, complete sdktypes.SessionCallAttemptComplete) error {
	return translateError(db.transaction(ctx, func(tx *tx) error {
		cdata, err := scheme.CompressJSON(complete)
		if err != nil {
			return fmt.Errorf("marshal session call attempt complete: %w", err)
		}

		r := scheme.SessionCallAttempt{
			CompressedComplete: cdata,
		}

		if res := tx.db.Model(&r).Where("session_id = ? AND seq = ? AND attempt = ?", sessionID.UUIDValue(), seq, attempt).Updates(r); res.Error != nil {
			return res.Error
		} else if res.RowsAffected == 0 {
			return sdkerrors.ErrNotFound
		}

		return addSessionLogRecord(tx.db, sessionID.UUIDValue(), sdktypes.NewCallAttemptCompleteSessionLogRecord(complete).WithProcessID(fixtures.ProcessID()))
	}))
}

// attempt = -1: latest.
// attempt >= 0: specific attempt.
func (db *gormdb) GetSessionCallAttemptResult(ctx context.Context, sessionID sdktypes.SessionID, seq uint32, attempt int64) (sdktypes.SessionCallAttemptResult, error) {
	q := db.db.Where("session_id = ? AND seq = ?", sessionID.UUIDValue(), seq)

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

	if r.Complete == nil && r.CompressedComplete == nil {
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
		Where("session_id = ? AND seq = ?", sessionID.UUIDValue(), seq).
		Count(&n).
		Error; err != nil {
		return 0, err
	}

	return uint32(n), nil
}
