package dbgorm

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	printSessionLogRecordType   = "print"
	stateSessionLogRecordType   = "state"
	stopSessionLogRecordType    = "stop_request"
	outcomeSessionLogRecordType = "outcome"
)

func (gdb *gormdb) createSession(ctx context.Context, session *scheme.Session) error {
	logr, err := toSessionLogRecord(session.SessionID, sdktypes.NewStateSessionLogRecord(kittehs.Now(), sdktypes.NewSessionStateCreated()))
	if err != nil {
		return err
	}

	return translateError(gdb.writeTransaction(ctx, func(tx *gormdb) error {
		if err := tx.writer.Create(session).Error; err != nil {
			return err
		}
		return createLogRecord(tx.writer, ctx, logr, stateSessionLogRecordType)
	}))
}

func (gdb *gormdb) deleteSession(ctx context.Context, sessionID uuid.UUID) error {
	return gdb.writeTransaction(ctx, func(tx *gormdb) error {
		var session scheme.Session
		err := tx.writer.WithContext(ctx).Where("session_id = ?", sessionID).First(&session).Error
		if err != nil {
			return err
		}

		err = tx.decDeploymentStats(ctx, session)
		if err != nil {
			return err
		}

		return tx.writer.Delete(&session).Error
	})
}

func (gdb *gormdb) updateSessionState(ctx context.Context, sessionID uuid.UUID, state sdktypes.SessionState) error {
	sessionStateUpdate := map[string]any{"current_state_type": int(state.Type().ToProto()), "updated_at": kittehs.Now()}
	logr, err := toSessionLogRecord(sessionID, sdktypes.NewStateSessionLogRecord(kittehs.Now(), state))
	if err != nil {
		return err
	}

	return gdb.writeTransaction(ctx, func(tx *gormdb) error {
		var session scheme.Session
		if err := tx.writer.First(&session, "session_id = ?", sessionID).Error; err != nil {
			return err
		}

		if err := tx.writer.Model(&scheme.Session{SessionID: sessionID}).Updates(sessionStateUpdate).Error; err != nil {
			return err
		}

		oldStateType := session.CurrentStateType
		newStateType := int(state.Type().ToProto())
		runningState := int(sdktypes.SessionStateTypeRunning.ToProto())

		if oldStateType <= runningState && newStateType > runningState && session.DeploymentID != nil {
			if err := gdb.incDeploymentStats(tx, *session.DeploymentID, newStateType); err != nil {
				return err
			}
		}

		return createLogRecord(tx.writer, ctx, logr, stateSessionLogRecordType)
	})
}

func (gdb *gormdb) getSession(ctx context.Context, sessionID uuid.UUID) (*scheme.Session, error) {
	return getOne[scheme.Session](gdb.reader.WithContext(ctx), "session_id = ?", sessionID)
}

func (gdb *gormdb) listSessions(ctx context.Context, f sdkservices.ListSessionsFilter) ([]scheme.Session, int64, error) {
	q := gdb.reader.WithContext(ctx)

	q = withProjectID(q, "sessions", f.ProjectID)

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
	return createLogRecord(gdb.writer, ctx, logr, typ)
}

func (gdb *gormdb) getSessionLogRecords(ctx context.Context, filter sdkservices.SessionLogRecordsFilter) (logs []scheme.SessionLogRecord, n int64, err error) {
	sessionID := filter.SessionID.UUIDValue()

	if err := gdb.writeTransaction(ctx, func(tx *gormdb) error {
		q := tx.reader.Where("session_id = ?", sessionID)

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
			specific(sdktypes.OutcomeSessionLogRecordType, outcomeSessionLogRecordType)

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
		IsDurable:        session.IsDurable(),
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

func (db *gormdb) AddSessionPrint(ctx context.Context, sessionID sdktypes.SessionID, v sdktypes.Value, callSeq uint32) error {
	logr, err := toSessionLogRecord(sessionID.UUIDValue(), sdktypes.NewPrintSessionLogRecord(kittehs.Now(), v, callSeq))
	if err != nil {
		return err
	}
	return translateError(db.addSessionLogRecord(ctx, logr, printSessionLogRecordType))
}

func (db *gormdb) AddSessionStopRequest(ctx context.Context, sessionID sdktypes.SessionID, reason string) error {
	logr, err := toSessionLogRecord(sessionID.UUIDValue(), sdktypes.NewStopRequestSessionLogRecord(kittehs.Now(), reason))
	if err != nil {
		return err
	}
	return translateError(db.addSessionLogRecord(ctx, logr, stopSessionLogRecordType))
}

func (db *gormdb) AddSessionOutcome(ctx context.Context, sessionID sdktypes.SessionID, v sdktypes.Value) error {
	logr, err := toSessionLogRecord(sessionID.UUIDValue(), sdktypes.NewOutcomeSessionLogRecord(kittehs.Now(), v))
	if err != nil {
		return err
	}
	return translateError(db.addSessionLogRecord(ctx, logr, outcomeSessionLogRecordType))
}

func (db *gormdb) GetSessionLog(ctx context.Context, filter sdkservices.SessionLogRecordsFilter) (*sdkservices.GetLogResults, error) {
	rs, n, err := db.getSessionLogRecords(ctx, filter)
	if err != nil {
		return nil, translateError(err)
	}

	prs, err := kittehs.TransformError(rs, scheme.ParseSessionLogRecord)
	if err != nil {
		return nil, err
	}

	nextPageToken := ""
	if len(rs) == int(filter.PageSize) && len(rs) > 0 {
		nextPageToken = strconv.FormatUint(rs[len(rs)-1].Seq, 10)
	}

	return &sdkservices.GetLogResults{Records: prs, PaginationResult: sdktypes.PaginationResult{TotalCount: n, NextPageToken: nextPageToken}}, err
}
