package dbgorm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
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

const (
	printSessionLogRecordType               = "print"
	stateSessionLogRecordType               = "state"
	stopSessionLogRecordType                = "stop_request"
	callSpecSessionLogRecordType            = "call_spec"
	callAttemptStartSessionLogRecordType    = "call_attempt_start"
	callAttemptCompleteSessionLogRecordType = "call_attempt_complete"
)

func (gdb *gormdb) createSession(ctx context.Context, session *scheme.Session) error {
	logr, err := toSessionLogRecord(session.SessionID, sdktypes.NewStateSessionLogRecord(sdktypes.NewSessionStateCreated()))
	if err != nil {
		return err
	}

	return translateError(gdb.transaction(ctx, func(tx *tx) error {
		if err := tx.db.Create(session).Error; err != nil {
			return err
		}
		return createLogRecord(tx.db, ctx, logr, stateSessionLogRecordType)
	}))
}

func (gdb *gormdb) deleteSession(ctx context.Context, sessionID uuid.UUID) error {
	return gdb.db.WithContext(ctx).Delete(&scheme.Session{SessionID: sessionID}).Error
}

func (gdb *gormdb) updateSessionState(ctx context.Context, sessionID uuid.UUID, state sdktypes.SessionState) error {
	sessionStateUpdate := map[string]any{"current_state_type": int(state.Type().ToProto()), "updated_at": kittehs.Now()}
	logr, err := toSessionLogRecord(sessionID, sdktypes.NewStateSessionLogRecord(state))
	if err != nil {
		return err
	}

	return gdb.transaction(ctx, func(tx *tx) error {
		if err := tx.db.Model(&scheme.Session{SessionID: sessionID}).Updates(sessionStateUpdate).Error; err != nil {
			return err
		}
		return createLogRecord(tx.db, ctx, logr, stateSessionLogRecordType)
	})
}

func (gdb *gormdb) getSession(ctx context.Context, sessionID uuid.UUID) (*scheme.Session, error) {
	return getOne[scheme.Session](gdb.db.WithContext(ctx), "session_id = ?", sessionID)
}

func (gdb *gormdb) listSessions(ctx context.Context, f sdkservices.ListSessionsFilter) ([]scheme.Session, int64, error) {
	q := gdb.db.WithContext(ctx)

	q = withProjectID(q, "", f.ProjectID)

	q = withProjectOrgID(q, f.OrgID, "sessions")

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

	var n int64
	err := q.Model(&scheme.Session{}).Count(&n).Error
	if err != nil {
		return nil, 0, err
	}

	if f.CountOnly {
		return nil, n, err
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

	var rs []scheme.Session
	// Double order in case we have two rows with the same created_at, then solve order by session_id
	if err := q.
		Order(clause.OrderByColumn{Column: clause.Column{Name: "sessions.created_at"}, Desc: true}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "session_id"}, Desc: true}).
		Omit("inputs").
		Find(&rs).Error; err != nil {
		return nil, 0, err
	}

	return rs, n, nil
}

// --- log records ---
func createLogRecord(db *gorm.DB, ctx context.Context, logr *scheme.SessionLogRecord, typ string) error {
	logr.Seq = uint64(time.Now().UnixMicro())
	logr.Type = typ
	return db.WithContext(ctx).Create(logr).Error
}

func (gdb *gormdb) addSessionLogRecord(ctx context.Context, logr *scheme.SessionLogRecord, typ string) error {
	return createLogRecord(gdb.db, ctx, logr, typ)
}

func (gdb *gormdb) getSessionLogRecords(ctx context.Context, filter sdkservices.ListSessionLogRecordsFilter) (logs []scheme.SessionLogRecord, n int64, err error) {
	sessionID := filter.SessionID.UUIDValue()

	if err := gdb.transaction(ctx, func(tx *tx) error {
		q := tx.db.Where("session_id = ?", sessionID)

		if err := q.Model(&scheme.SessionLogRecord{}).Count(&n).Error; err != nil {
			return err
		}

		if filter.PageSize != 0 {
			q = q.Limit(int(filter.PageSize))
		}

		if filter.Skip != 0 {
			q = q.Offset(int(filter.Skip))
		}

		if filter.PageToken != "" {
			if filter.Ascending {
				q = q.Where("seq > ?", filter.PageToken)
			} else {
				q = q.Where("seq < ?", filter.PageToken)
			}
		}

		if types := filter.Types; types != 0 {
			var qtypes []string

			specific := func(t sdktypes.SessionLogRecordType, name string) {
				if types&t != 0 {
					qtypes = append(qtypes, name)
				}
			}

			specific(sdktypes.PrintSessionLogRecordType, printSessionLogRecordType)
			specific(sdktypes.StateSessionLogRecordType, stateSessionLogRecordType)
			specific(sdktypes.StopRequestSessionLogRecordType, stopSessionLogRecordType)
			specific(sdktypes.CallSpecSessionLogRecordType, callSpecSessionLogRecordType)
			specific(sdktypes.CallAttemptStartSessionLogRecordType, callAttemptStartSessionLogRecordType)
			specific(sdktypes.CallAttemptCompleteSessionLogRecordType, callAttemptCompleteSessionLogRecordType)

			q = q.Where("type in (?)", qtypes)
		}

		// Default is desc order
		q = q.Order(clause.OrderByColumn{Column: clause.Column{Name: "seq"}, Desc: !filter.Ascending})
		return q.Find(&logs).Error
	}); err != nil {
		return nil, n, err
	}
	return logs, n, nil
}

// --- session calls ---
func (gdb *gormdb) createSessionCall(ctx context.Context, sessionID uuid.UUID, spec sdktypes.SessionCallSpec) error {
	jsonSpec, err := json.Marshal(spec)
	if err != nil {
		return fmt.Errorf("marshal session call: %w", err)
	}

	logr, err := toSessionLogRecord(sessionID, sdktypes.NewCallSpecSessionLogRecord(spec))
	if err != nil {
		return err
	}

	callSpec := scheme.SessionCallSpec{
		SessionID: sessionID,
		Seq:       spec.Seq(),
		Data:      jsonSpec,
	}

	return gdb.transaction(ctx, func(tx *tx) error {
		if err := tx.db.Create(&callSpec).Error; err != nil {
			return err
		}
		return createLogRecord(tx.db, ctx, logr, callSpecSessionLogRecordType)
	})
}

func (gdb *gormdb) getSessionCallSpec(ctx context.Context, sessionID uuid.UUID, seq uint32) (*scheme.SessionCallSpec, error) {
	var r scheme.SessionCallSpec
	if err := gdb.db.WithContext(ctx).Where("session_id = ?", sessionID).Where("seq = ?", seq).First(&r).Error; err != nil {
		return nil, err
	}
	return &r, nil
}

func countCallAttempts(db *gorm.DB, sessionID uuid.UUID, seq uint32) (uint32, error) {
	var n int64
	if err := db.Model(&scheme.SessionCallAttempt{}).
		Where("session_id = ? AND seq = ?", sessionID, seq).Count(&n).Error; err != nil {
		return 0, err
	}

	return uint32(n), nil
}

func (gdb *gormdb) startSessionCallAttempt(ctx context.Context, sessionID uuid.UUID, seq uint32) (attempt uint32, err error) {
	err = gdb.transaction(ctx, func(tx *tx) error {
		if attempt, err = countCallAttempts(tx.db, sessionID, seq); err != nil {
			return err
		}

		callAttemptStart := kittehs.Must1(sdktypes.SessionCallAttemptStartFromProto(&sdktypes.SessionCallAttemptStartPB{
			StartedAt: timestamppb.Now(),
			Num:       attempt,
		}))
		callAttemptStartJson, err := json.Marshal(callAttemptStart)
		if err != nil {
			return err
		}
		logr, err := toSessionLogRecord(sessionID, sdktypes.NewCallAttemptStartSessionLogRecord(callAttemptStart))
		if err != nil {
			return err
		}

		callAttempt := scheme.SessionCallAttempt{
			SessionID: sessionID,
			Seq:       seq,
			Attempt:   attempt,
			Start:     callAttemptStartJson,
		}

		if err := tx.db.Create(&callAttempt).Error; err != nil {
			return err
		}

		return createLogRecord(tx.db, ctx, logr, callAttemptStartSessionLogRecordType)
	})
	return
}

func (gdb *gormdb) completeSessionCallAttempt(ctx context.Context, sessionID uuid.UUID, seq, attempt uint32, complete sdktypes.SessionCallAttemptComplete) error {
	logr, err := toSessionLogRecord(sessionID, sdktypes.NewCallAttemptCompleteSessionLogRecord(complete))
	if err != nil {
		return err
	}

	json, err := json.Marshal(complete)
	if err != nil {
		return fmt.Errorf("marshal session call attempt complete: %w", err)
	}
	r := scheme.SessionCallAttempt{Complete: json}

	return gdb.transaction(ctx, func(tx *tx) error {
		if res := tx.db.Model(&r).Where("session_id = ? AND seq = ? AND attempt = ?", sessionID, seq, attempt).Updates(r); res.Error != nil {
			return res.Error
		} else if res.RowsAffected == 0 {
			return sdkerrors.ErrNotFound
		}
		return createLogRecord(tx.db, ctx, logr, callAttemptCompleteSessionLogRecordType)
	})
}

// attempt legend: -1 for latest, >= 0 for specific attempt.
func (gdb *gormdb) getSessionCallAttemptResult(ctx context.Context, sessionID uuid.UUID, seq uint32, attempt int64) (*scheme.SessionCallAttempt, error) {
	var r scheme.SessionCallAttempt

	if err := gdb.transaction(ctx, func(tx *tx) error {
		q := tx.db.Where("session_id = ? AND seq = ?", sessionID, seq)

		if attempt == -1 {
			q = q.Order(clause.OrderByColumn{Column: clause.Column{Name: "attempt"}, Desc: true})
		} else if attempt >= 0 {
			q = q.Where("attempt = ?", attempt)
		} else {
			return sdkerrors.NewInvalidArgumentError("attempt must be either -1 or >= 0, got %d", attempt)
		}

		if err := q.First(&r).Error; err != nil {
			return err
		}
		if r.Complete == nil {
			return sdkerrors.ErrNotFound
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &r, nil
}

func (db *gormdb) CreateSession(ctx context.Context, session sdktypes.Session) error {
	if err := session.Strict(); err != nil {
		return err
	}

	s := scheme.Session{
		Base:             based(ctx),
		ProjectID:        session.ProjectID().UUIDValue(),
		SessionID:        session.ID().UUIDValue(),
		BuildID:          session.BuildID().UUIDValue(),
		DeploymentID:     uuidPtrOrNil(session.DeploymentID()),
		EventID:          uuidPtrOrNil(session.EventID()),
		Entrypoint:       session.EntryPoint().CanonicalString(),
		CurrentStateType: int(sdktypes.SessionStateTypeCreated.ToProto()),
		Inputs:           kittehs.Must1(json.Marshal(session.Inputs())),
		Memo:             kittehs.Must1(json.Marshal(session.Memo())),
	}
	return translateError(db.createSession(ctx, &s))
}

func (db *gormdb) DeleteSession(ctx context.Context, sessionID sdktypes.SessionID) error {
	return translateError(db.deleteSession(ctx, sessionID.UUIDValue()))
}

func (db *gormdb) UpdateSessionState(ctx context.Context, sessionID sdktypes.SessionID, state sdktypes.SessionState) error {
	return translateError(db.updateSessionState(ctx, sessionID.UUIDValue(), state))
}

func (db *gormdb) GetSession(ctx context.Context, id sdktypes.SessionID) (sdktypes.Session, error) {
	s, err := db.getSession(ctx, id.UUIDValue())
	if s == nil || err != nil {
		return sdktypes.InvalidSession, translateError(err)
	}
	return scheme.ParseSession(*s)
}

func (db *gormdb) ListSessions(ctx context.Context, f sdkservices.ListSessionsFilter) (*sdkservices.ListSessionResult, error) {
	rs, cnt, err := db.listSessions(ctx, f)
	if err != nil {
		return nil, translateError(err)
	}

	sessions, err := kittehs.TransformError(rs, scheme.ParseSession)
	if err != nil {
		return nil, err
	}

	// Only if we have a full page, there might be more sessions
	nextPageToken := ""
	if len(sessions) == int(f.PageSize) && len(sessions) > 0 {
		nextPageToken = sessions[len(sessions)-1].ID().UUIDValue().String()
	}

	return &sdkservices.ListSessionResult{
		Sessions:         sessions,
		PaginationResult: sdktypes.PaginationResult{TotalCount: cnt, NextPageToken: nextPageToken},
	}, nil
}

// --- log records funcs ---
func toSessionLogRecord(sessionID uuid.UUID, logr sdktypes.SessionLogRecord) (*scheme.SessionLogRecord, error) {
	logr = logr.WithProcessID(fixtures.ProcessID())
	logRecordData, err := json.Marshal(logr)
	if err != nil {
		return nil, fmt.Errorf("marshal session log record: %w", err)
	}

	return &scheme.SessionLogRecord{SessionID: sessionID, Data: logRecordData}, nil
}

func (db *gormdb) AddSessionPrint(ctx context.Context, sessionID sdktypes.SessionID, print string) error {
	logr, err := toSessionLogRecord(sessionID.UUIDValue(), sdktypes.NewPrintSessionLogRecord(print))
	if err != nil {
		return err
	}
	return translateError(db.addSessionLogRecord(ctx, logr, printSessionLogRecordType))
}

func (db *gormdb) AddSessionStopRequest(ctx context.Context, sessionID sdktypes.SessionID, reason string) error {
	logr, err := toSessionLogRecord(sessionID.UUIDValue(), sdktypes.NewStopRequestSessionLogRecord(reason))
	if err != nil {
		return err
	}
	return translateError(db.addSessionLogRecord(ctx, logr, stopSessionLogRecordType))
}

func (db *gormdb) GetSessionLog(ctx context.Context, filter sdkservices.ListSessionLogRecordsFilter) (*sdkservices.GetLogResults, error) {
	rs, n, err := db.getSessionLogRecords(ctx, filter)
	if err != nil {
		return nil, translateError(err)
	}

	prs, err := kittehs.TransformError(rs, scheme.ParseSessionLogRecord)
	log := sdktypes.NewSessionLog(prs)

	nextPageToken := ""
	if len(rs) == int(filter.PageSize) && len(rs) > 0 {
		nextPageToken = fmt.Sprintf("%d", rs[len(rs)-1].Seq)
	}

	return &sdkservices.GetLogResults{Log: log, PaginationResult: sdktypes.PaginationResult{TotalCount: n, NextPageToken: nextPageToken}}, err
}

// --- session call funcs ---
func (db *gormdb) CreateSessionCall(ctx context.Context, sessionID sdktypes.SessionID, spec sdktypes.SessionCallSpec) error {
	return translateError(db.createSessionCall(ctx, sessionID.UUIDValue(), spec))
}

func (db *gormdb) GetSessionCallSpec(ctx context.Context, sessionID sdktypes.SessionID, seq uint32) (sdktypes.SessionCallSpec, error) {
	r, err := db.getSessionCallSpec(ctx, sessionID.UUIDValue(), seq)
	if r == nil || err != nil {
		return sdktypes.InvalidSessionCallSpec, translateError(err)
	}
	return scheme.ParseSessionCallSpec(*r)
}

func (db *gormdb) StartSessionCallAttempt(ctx context.Context, sessionID sdktypes.SessionID, seq uint32) (attempt uint32, err error) {
	attempt, err = db.startSessionCallAttempt(ctx, sessionID.UUIDValue(), seq)
	return attempt, translateError(err)
}

func (db *gormdb) CompleteSessionCallAttempt(ctx context.Context, sessionID sdktypes.SessionID, seq, attempt uint32, complete sdktypes.SessionCallAttemptComplete) error {
	return translateError(db.completeSessionCallAttempt(ctx, sessionID.UUIDValue(), seq, attempt, complete))
}

// attempt legend: -1 for latest, >= 0 for specific attempt.
func (db *gormdb) GetSessionCallAttemptResult(ctx context.Context, sessionID sdktypes.SessionID, seq uint32, attempt int64) (sdktypes.SessionCallAttemptResult, error) {
	rs, err := db.getSessionCallAttemptResult(ctx, sessionID.UUIDValue(), seq, attempt)
	if err != nil {
		return sdktypes.InvalidSessionCallAttemptResult, translateError(err)
	}
	complete, err := scheme.ParseSessionCallAttemptComplete(*rs)
	if err != nil {
		return sdktypes.InvalidSessionCallAttemptResult, err
	}

	return complete.Result(), nil
}
