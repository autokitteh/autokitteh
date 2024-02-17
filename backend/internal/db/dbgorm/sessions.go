package dbgorm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"go.autokitteh.dev/autokitteh/backend/internal/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (db *gormdb) GetSession(ctx context.Context, sessionID sdktypes.SessionID) (sdktypes.Session, error) {
	return get(db.db, ctx, scheme.ParseSession, "session_id = ?", sessionID.String())
}

func (db *gormdb) GetSessionLog(ctx context.Context, sessionID sdktypes.SessionID) (sdktypes.SessionLog, error) {
	var rs []scheme.SessionLogRecord

	if err := db.db.WithContext(ctx).Where("session_id = ?", sessionID.String()).Find(&rs).Error; err != nil {
		return nil, err
	}

	prs, err := kittehs.TransformError(rs, scheme.ParseSessionLogRecord)

	return sdktypes.NewSessionLog(prs), err
}

func addSessionLogRecord(tx *gorm.DB, sessionID sdktypes.SessionID, logr sdktypes.SessionLogRecord) error {
	jsonData, err := json.Marshal(logr)
	if err != nil {
		return fmt.Errorf("marshal session log record: %w", err)
	}

	r := scheme.SessionLogRecord{
		SessionID: sessionID.String(),
		Data:      jsonData,
	}

	return tx.Create(&r).Error
}

func (db *gormdb) CreateSession(ctx context.Context, session sdktypes.Session) error {
	return translateError(db.transaction(ctx, func(tx *tx) error {
		now := time.Now()

		r := scheme.Session{
			SessionID:        sdktypes.GetSessionID(session).String(),
			DeploymentID:     sdktypes.GetSessionDeploymentID(session).String(),
			EventID:          sdktypes.GetSessionEventID(session).String(),
			Entrypoint:       sdktypes.GetCodeLocationCanonicalString(sdktypes.GetSessionEntryPoint(session)),
			CurrentStateType: int(sdktypes.CreatedSessionStateType),
			Inputs:           kittehs.Must1(json.Marshal(sdktypes.GetSessionInputs(session))),
			CreatedAt:        now,
			UpdatedAt:        now,
		}

		if err := tx.db.WithContext(ctx).Create(&r).Error; err != nil {
			return translateError(err)
		}

		return addSessionLogRecord(tx.db, sdktypes.GetSessionID(session), sdktypes.NewStateSessionLogRecord(
			kittehs.Must1(sdktypes.WrapSessionState(sdktypes.NewCreatedSessionState()).Update(
				func(pb *sdktypes.SessionStatePB) { pb.T = timestamppb.Now() },
			)),
		))
	}))
}

func (db *gormdb) UpdateSessionState(ctx context.Context, sessionID sdktypes.SessionID, state sdktypes.SessionState) error {
	return translateError(db.transaction(ctx, func(tx *tx) error {
		r := scheme.Session{
			CurrentStateType: int(sdktypes.GetSessionStateType(state).ToProto()),
			UpdatedAt:        time.Now(),
		}

		if res := tx.db.Model(&r).Where("session_id = ?", sessionID.String()).Updates(r); res.Error != nil {
			return res.Error
		} else if res.RowsAffected == 0 {
			return sdkerrors.ErrNotFound
		}

		return addSessionLogRecord(tx.db, sessionID, sdktypes.NewSessionLogState(
			kittehs.Must1(state.Update(
				func(pb *sdktypes.SessionStatePB) { pb.T = timestamppb.Now() },
			)),
		))
	}))
}

func (db *gormdb) AddSessionPrint(ctx context.Context, sessionID sdktypes.SessionID, print string) error {
	return addSessionLogRecord(db.db, sessionID, sdktypes.NewSessionLogPrint(print))
}

func (db *gormdb) ListSessions(ctx context.Context, f sdkservices.ListSessionsFilter) ([]sdktypes.Session, int, error) {
	var rs []scheme.Session

	q := db.db.WithContext(ctx)

	if f.DeploymentID != nil {
		q = q.Where("deployment_id = ?", f.DeploymentID.String())
	}

	if f.EventID != nil {
		q = q.Where("event_id = ?", f.EventID.String())
	}

	if f.StateType != sdktypes.UnspecifiedSessionStateType {
		q = q.Where("current_state_type = ?", f.StateType.ToProto())
	}

	if f.CountOnly {
		var n int64
		err := q.Model(&scheme.Session{}).Count(&n).Error
		return nil, int(n), err
	}

	if err := q.Order("created_at desc").Find(&rs).Error; err != nil {
		return nil, 0, translateError(err)
	}

	xs, err := kittehs.TransformError(rs, scheme.ParseSession)
	return xs, len(xs), err
}

func (db *gormdb) CreateSessionCall(ctx context.Context, sessionID sdktypes.SessionID, spec sdktypes.SessionCallSpec) error {
	jsonSpec, err := json.Marshal(spec)
	if err != nil {
		return fmt.Errorf("marshal session call: %w", err)
	}

	return translateError(db.transaction(ctx, func(tx *tx) error {
		r := scheme.SessionCallSpec{
			SessionID: sessionID.String(),
			Seq:       sdktypes.GetSessionCallSpecSeq(spec),
			Data:      jsonSpec,
		}

		if err := tx.db.Create(&r).Error; err != nil {
			return err
		}

		return addSessionLogRecord(tx.db, sessionID, sdktypes.NewSessionLogCallSpec(spec))
	}))
}

func (db *gormdb) GetSessionCallSpec(ctx context.Context, sessionID sdktypes.SessionID, seq uint32) (sdktypes.SessionCallSpec, error) {
	var r scheme.SessionCallSpec
	if err := db.db.
		Where("session_id = ?", sessionID.String()).
		Where("seq = ?", seq).
		First(&r).
		Error; err != nil {
		return nil, translateError(err)
	}

	spec, err := scheme.ParseSessionCallSpec(r)
	if err != nil {
		return nil, err
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

		r := scheme.SessionCallAttempt{
			SessionID: sessionID.String(),
			Seq:       seq,
			Attempt:   attempt,
			Start:     json,
		}

		if err := tx.db.Create(&r).Error; err != nil {
			return err
		}

		return addSessionLogRecord(tx.db, sessionID, sdktypes.NewSessionLogCallAttemptStart(obj))
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

		return addSessionLogRecord(tx.db, sessionID, sdktypes.NewSessionLogCallAttemptComplete(complete))
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
		return nil, fmt.Errorf("attempt must be either -1 or >= 0: %w", sdkerrors.ErrInvalidArgument)
	}

	var r scheme.SessionCallAttempt
	if err := q.First(&r).Error; err != nil {
		return nil, translateError(err)
	}

	if r.Complete == nil {
		return nil, sdkerrors.ErrNotFound
	}

	complete, err := scheme.ParseSessionCallAttemptComplete(r)
	if err != nil {
		return nil, err
	}

	return sdktypes.GetSessionCallAttemptCompleteResult(complete), nil
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
