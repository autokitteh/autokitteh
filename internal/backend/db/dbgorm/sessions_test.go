package dbgorm

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (f *dbFixture) createSessionsAndAssert(t *testing.T, sessions ...scheme.Session) {
	for _, session := range sessions {
		assert.NoError(t, f.gormdb.createSession(f.ctx, &session))
		findAndAssertOne(t, f, session, "session_id = ?", session.SessionID)
	}
}

func (f *dbFixture) assertSessionsLogRecordsDeleted(t *testing.T, records ...scheme.SessionLogRecord) {
	for _, record := range records {
		assertDeleted(t, f, record)
	}
}

func (f *dbFixture) listSessionsAndAssert(t *testing.T, expected int64) []scheme.Session {
	flt := sdkservices.ListSessionsFilter{
		CountOnly: false,
		StateType: sdktypes.SessionStateTypeUnspecified, // fetch all sessions
	}

	sessions, cnt, err := f.gormdb.listSessions(f.ctx, flt)
	require.NoError(t, err)
	assert.Equal(t, expected, cnt)
	require.Equal(t, expected, int64(len(sessions)))

	return sessions
}

func (f *dbFixture) assertSessionsDeleted(t *testing.T, sessions ...scheme.Session) {
	for _, session := range sessions {
		assertDeleted(t, f, session)
	}
}

func preSessionTest(t *testing.T) (*dbFixture, scheme.Project, scheme.Build) {
	f := newDBFixture()
	f.listSessionsAndAssert(t, 0) // no sessions
	findAndAssertCount[scheme.SessionLogRecord](t, f, 0, "")

	p, b := f.createProjectBuild(t)

	return f, p, b
}

func testLastLogRecord(t *testing.T, f *dbFixture, numRecords int, sessionID uuid.UUID, lr sdktypes.SessionLogRecord) {
	sid := sdktypes.NewIDFromUUID[sdktypes.SessionID](sessionID)
	logs, _, err := f.gormdb.getSessionLogRecords(f.ctx, sdkservices.ListSessionLogRecordsFilter{SessionID: sid, PaginationRequest: sdktypes.PaginationRequest{Ascending: false}})
	assert.NoError(t, err)
	assert.Equal(t, numRecords, len(logs))
	for _, r := range logs {
		assert.Equal(t, sessionID, r.SessionID)
	}

	l, err := scheme.ParseSessionLogRecord(logs[0]) // last log
	assert.NoError(t, err)

	l = l.WithoutTimestamp()
	lr = lr.WithProcessID(fixtures.ProcessID()).WithoutTimestamp()
	assert.Equal(t, l, lr)
}

func assertSessionLogRecordsEqual(t *testing.T, a, b sdktypes.SessionLogRecord) {
	a = a.WithProcessID(fixtures.ProcessID()).WithoutTimestamp()
	b = b.WithProcessID(fixtures.ProcessID()).WithoutTimestamp()
	assert.Equal(t, a, b)
}

func TestCreateSession(t *testing.T) {
	f, p, b := preSessionTest(t)

	// test createSession
	s := f.newSession(sdktypes.SessionStateTypeCompleted, p, b)
	f.createSessionsAndAssert(t, s) // all session assets are optional and will be set to nil

	// test getSessionLogRecords and ensure that session logs contain the only CREATED record
	testLastLogRecord(t, f, 1, s.SessionID, sdktypes.NewStateSessionLogRecord(sdktypes.NewSessionStateCreated()))
}

func TestCreateSessionForeignKeys(t *testing.T) {
	// check session creation if foreign keys are not nil
	f, p, b := preSessionTest(t)

	// test with existing assets
	d := f.newDeployment(b, p)
	evt := f.newEvent(p)
	s := f.newSession(sdktypes.SessionStateTypeCompleted, d, b, p, evt)

	f.createDeploymentsAndAssert(t, d)
	f.createEventsAndAssert(t, evt)
	f.createSessionsAndAssert(t, s)

	// negative test with non-existing assets
	// use existing user owned projectID as fakeID to pass user check

	s2 := f.newSession(sdktypes.SessionStateTypeCompleted)

	s2.BuildID = p.ProjectID // no such buildID, since it's a projectID
	assert.ErrorIs(t, f.gormdb.createSession(f.ctx, &s2), gorm.ErrForeignKeyViolated)

	s2.DeploymentID = &p.ProjectID // no such deploymentID, since it's a projectID
	assert.ErrorIs(t, f.gormdb.createSession(f.ctx, &s2), gorm.ErrForeignKeyViolated)
	s2.DeploymentID = nil

	s2.EventID = &p.ProjectID // no such eventID, since it's a projectID
	assert.ErrorIs(t, f.gormdb.createSession(f.ctx, &s2), gorm.ErrForeignKeyViolated)
	s2.EventID = nil
}

func TestGetSession(t *testing.T) {
	f, p, b := preSessionTest(t)

	s := f.newSession(sdktypes.SessionStateTypeCompleted, p, b)
	f.createSessionsAndAssert(t, s)

	// check getSession
	session, err := f.gormdb.getSession(f.ctx, s.SessionID)
	assert.NoError(t, err)
	resetTimes(session)
	assert.Equal(t, s, *session)

	// check that after deleteSession it's not found
	assert.NoError(t, f.gormdb.deleteSession(f.ctx, s.SessionID))
	_, err = f.gormdb.getSession(f.ctx, s.SessionID)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestListSessions(t *testing.T) {
	f, p, b := preSessionTest(t)

	s := f.newSession(sdktypes.SessionStateTypeCompleted, p, b)
	f.createSessionsAndAssert(t, s)

	sessions := f.listSessionsAndAssert(t, 1)
	s.Inputs = nil
	resetTimes(&sessions[0])
	assert.Equal(t, s, sessions[0])

	// deleteSession and ensure that listSessions is empty
	assert.NoError(t, f.gormdb.deleteSession(f.ctx, s.SessionID))
	f.listSessionsAndAssert(t, 0)
}

func TestListSessionsNoSessions(t *testing.T) {
	f, _, _ := preSessionTest(t)

	sessions := f.listSessionsAndAssert(t, 0)
	assert.Equal(t, sessions, []scheme.Session{})
}

func TestListSessionsCountOnly(t *testing.T) {
	f, p, b := preSessionTest(t)

	s := f.newSession(sdktypes.SessionStateTypeCompleted, p, b)
	f.createSessionsAndAssert(t, s)

	flt := sdkservices.ListSessionsFilter{
		CountOnly: true,
		StateType: sdktypes.SessionStateTypeUnspecified, // fetch all sessions
	}

	sessions, cnt, err := f.gormdb.listSessions(f.ctx, flt)
	require.NoError(t, err)
	require.Equal(t, cnt, int64(1))
	require.Nil(t, sessions)
}

func TestListSessionsNoSessionsCountOnly(t *testing.T) {
	f, _, _ := preSessionTest(t)

	flt := sdkservices.ListSessionsFilter{
		CountOnly: true,
		StateType: sdktypes.SessionStateTypeUnspecified, // fetch all sessions
	}

	sessions, cnt, err := f.gormdb.listSessions(f.ctx, flt)
	require.NoError(t, err)
	require.Equal(t, cnt, int64(0))
	require.Nil(t, sessions)
}

func TestListPaginatedSession(t *testing.T) {
	f, p, b := preSessionTest(t)

	s := f.newSession(sdktypes.SessionStateTypeCompleted, p, b)
	f.createSessionsAndAssert(t, s)

	s = f.newSession(sdktypes.SessionStateTypeCompleted, p, b)
	f.createSessionsAndAssert(t, s)

	flt := sdkservices.ListSessionsFilter{
		CountOnly:         false,
		StateType:         sdktypes.SessionStateTypeUnspecified, // fetch all sessions
		PaginationRequest: sdktypes.PaginationRequest{PageSize: 1},
	}

	sessions, cnt, err := f.gormdb.listSessions(f.ctx, flt)
	require.NoError(t, err)
	require.Equal(t, cnt, int64(2))
	require.Len(t, sessions, 1)
}

func TestDeleteSession(t *testing.T) {
	f, p, b := preSessionTest(t)

	s := f.newSession(sdktypes.SessionStateTypeCompleted, p, b)
	f.createSessionsAndAssert(t, s)

	assert.NoError(t, f.gormdb.deleteSession(f.ctx, s.SessionID))
	f.assertSessionsDeleted(t, s)
}

func TestDeleteSessionLogRecords(t *testing.T) {
	f, p, b := preSessionTest(t)

	s := f.newSession(sdktypes.SessionStateTypeCompleted, p, b)
	f.createSessionsAndAssert(t, s)

	lr := f.newSessionLogRecord(s.SessionID)
	assert.NoError(t, f.gormdb.addSessionLogRecord(f.ctx, &lr, ""))

	assert.NoError(t, f.gormdb.deleteSession(f.ctx, s.SessionID))
	f.assertSessionsDeleted(t, s)
	f.assertSessionsLogRecordsDeleted(t, lr)
}

func TestDeleteSessionCallSpec(t *testing.T) {
	f, p, b := preSessionTest(t)

	// Create Session
	s := f.newSession(sdktypes.SessionStateTypeCompleted, p, b)
	f.createSessionsAndAssert(t, s)

	// Create SessionCallSpec
	cs := sdktypes.NewSessionCallSpec(sdktypes.InvalidValue, nil, nil, 1)
	assert.NoError(t, f.gormdb.createSessionCall(f.ctx, s.SessionID, cs))
	_, err := f.gormdb.getSessionCallSpec(f.ctx, s.SessionID, 1)
	assert.NoError(t, err)

	// Delete Session
	assert.NoError(t, f.gormdb.deleteSession(f.ctx, s.SessionID))

	f.assertSessionsDeleted(t, s)

	// Ensure that SessionCallSpec is deleted
	_, err = f.gormdb.getSessionCallSpec(f.ctx, s.SessionID, 1)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	// Session call implicitly create session log records
	_, n, err := f.gormdb.getSessionLogRecords(f.ctx, sdkservices.ListSessionLogRecordsFilter{SessionID: sdktypes.NewIDFromUUID[sdktypes.SessionID](s.SessionID)})
	assert.NoError(t, err)
	assert.Equal(t, n, int64(0))
}

func TestDeleteSessionCallAttempt(t *testing.T) {
	f, p, b := preSessionTest(t)

	// Create Session
	s := f.newSession(sdktypes.SessionStateTypeCompleted, p, b)
	f.createSessionsAndAssert(t, s)

	// Create SessionCallAttempt
	attempt, err := f.gormdb.startSessionCallAttempt(f.ctx, s.SessionID, 1)
	assert.NoError(t, err)

	// Delete Session
	assert.NoError(t, f.gormdb.deleteSession(f.ctx, s.SessionID))
	f.assertSessionsDeleted(t, s)

	// Ensure that session call attempt is deleted
	_, err = f.gormdb.getSessionCallAttemptResult(f.ctx, s.SessionID, 1, int64(attempt))
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	// Session call attempt create session log records
	_, n, err := f.gormdb.getSessionLogRecords(f.ctx, sdkservices.ListSessionLogRecordsFilter{SessionID: sdktypes.NewIDFromUUID[sdktypes.SessionID](s.SessionID)})
	assert.NoError(t, err)
	assert.Equal(t, n, int64(0))
}

/*
func TestDeleteSessionForeignKeys(t *testing.T) {
    // session is soft-deleted, so no need to check foreign keys meanwhile
}
*/

func TestAddSessionLogRecordUnexistingSession(t *testing.T) {
	f, p, b := preSessionTest(t)

	s := f.newSession(sdktypes.SessionStateTypeCompleted, p, b)
	logr := f.newSessionLogRecord(testDummyID) // invalid sessionID

	assert.ErrorIs(t, f.gormdb.addSessionLogRecord(f.ctx, &logr, ""), gorm.ErrForeignKeyViolated)

	f.createSessionsAndAssert(t, s) // will create session and session record as well

	// ensure that with valid sessionID log could be added
	logr.SessionID = s.SessionID
	assert.NoError(t, f.gormdb.addSessionLogRecord(f.ctx, &logr, ""))
}

func TestAddSessionPrintLogRecord(t *testing.T) {
	f, p, b := preSessionTest(t)

	s := f.newSession(sdktypes.SessionStateTypeCompleted, p, b)
	l := sdktypes.NewPrintSessionLogRecord("meow")
	logr, err := toSessionLogRecord(s.SessionID, l)
	assert.NoError(t, err)

	f.createSessionsAndAssert(t, s) // will create session and session record as well
	assert.NoError(t, f.gormdb.addSessionLogRecord(f.ctx, logr, ""))

	testLastLogRecord(t, f, 2, s.SessionID, l)
}

func TestSessionLogRecordListOrder(t *testing.T) {
	f, p, b := preSessionTest(t)

	s := f.newSession(sdktypes.SessionStateTypeCompleted, p, b)
	l := sdktypes.NewPrintSessionLogRecord("meow")
	logr, err := toSessionLogRecord(s.SessionID, l)
	assert.NoError(t, err)

	f.createSessionsAndAssert(t, s) // will create session and session record as well
	assert.NoError(t, f.gormdb.addSessionLogRecord(f.ctx, logr, ""))

	sid := sdktypes.NewIDFromUUID[sdktypes.SessionID](s.SessionID)

	tests := []struct {
		name  string
		asc   bool
		index int
	}{
		{
			name:  "ascending",
			asc:   true,
			index: 1,
		},
		{
			name:  "desc",
			asc:   false,
			index: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logs, _, err := f.gormdb.getSessionLogRecords(f.ctx, sdkservices.ListSessionLogRecordsFilter{SessionID: sid, PaginationRequest: sdktypes.PaginationRequest{Ascending: tt.asc}})
			assert.NoError(t, err)

			savedRecord, _ := scheme.ParseSessionLogRecord(logs[tt.index]) // last log
			assertSessionLogRecordsEqual(t, l, savedRecord)
		})
	}
}

func TestSessionLogRecordPageSizeAndTotalCount(t *testing.T) {
	f, p, b := preSessionTest(t)

	s := f.newSession(sdktypes.SessionStateTypeCompleted, p, b)
	l := sdktypes.NewPrintSessionLogRecord("meow")
	logr, err := toSessionLogRecord(s.SessionID, l)
	assert.NoError(t, err)

	f.createSessionsAndAssert(t, s) // will create session and session record as well
	assert.NoError(t, f.gormdb.addSessionLogRecord(f.ctx, logr, ""))

	sid := sdktypes.NewIDFromUUID[sdktypes.SessionID](s.SessionID)
	logs, n, err := f.gormdb.getSessionLogRecords(f.ctx,
		sdkservices.ListSessionLogRecordsFilter{
			SessionID:         sid,
			PaginationRequest: sdktypes.PaginationRequest{PageSize: 1},
		})

	assert.NoError(t, err)
	assert.Equal(t, len(logs), 1)
	assert.Equal(t, n, int64(2))
}

func TestSessionLogRecordSkipAll(t *testing.T) {
	f, p, b := preSessionTest(t)

	s := f.newSession(sdktypes.SessionStateTypeCompleted, p, b)
	l := sdktypes.NewPrintSessionLogRecord("meow")
	logr, err := toSessionLogRecord(s.SessionID, l)
	assert.NoError(t, err)

	f.createSessionsAndAssert(t, s) // will create session and session record as well
	assert.NoError(t, f.gormdb.addSessionLogRecord(f.ctx, logr, ""))

	sid := sdktypes.NewIDFromUUID[sdktypes.SessionID](s.SessionID)
	logs, n, err := f.gormdb.getSessionLogRecords(f.ctx,
		sdkservices.ListSessionLogRecordsFilter{
			SessionID:         sid,
			PaginationRequest: sdktypes.PaginationRequest{Skip: 2},
		})

	assert.NoError(t, err)
	assert.Equal(t, len(logs), 0)
	assert.Equal(t, n, int64(2))
}

func TestSessionLogRecordSkip(t *testing.T) {
	f, p, b := preSessionTest(t)

	s := f.newSession(sdktypes.SessionStateTypeCompleted, p, b)
	l := sdktypes.NewPrintSessionLogRecord("meow")
	logr, err := toSessionLogRecord(s.SessionID, l)
	assert.NoError(t, err)

	f.createSessionsAndAssert(t, s) // will create session and session record as well
	assert.NoError(t, f.gormdb.addSessionLogRecord(f.ctx, logr, ""))

	sid := sdktypes.NewIDFromUUID[sdktypes.SessionID](s.SessionID)
	logs, n, err := f.gormdb.getSessionLogRecords(f.ctx,
		sdkservices.ListSessionLogRecordsFilter{
			SessionID:         sid,
			PaginationRequest: sdktypes.PaginationRequest{Skip: 1},
		})

	assert.NoError(t, err)
	assert.Equal(t, len(logs), 1)
	assert.Equal(t, n, int64(2))
}

func TestSessionLogRecordNextPageTokenEmpty(t *testing.T) {
	f, p, b := preSessionTest(t)

	s := f.newSession(sdktypes.SessionStateTypeCompleted, p, b)
	l := sdktypes.NewPrintSessionLogRecord("meow")
	logr, err := toSessionLogRecord(s.SessionID, l)
	assert.NoError(t, err)

	f.createSessionsAndAssert(t, s) // will create session and session record as well
	assert.NoError(t, f.gormdb.addSessionLogRecord(f.ctx, logr, ""))

	sid := sdktypes.NewIDFromUUID[sdktypes.SessionID](s.SessionID)
	res, err := f.gormdb.GetSessionLog(context.Background(),
		sdkservices.ListSessionLogRecordsFilter{
			SessionID:         sid,
			PaginationRequest: sdktypes.PaginationRequest{},
		})

	assert.NoError(t, err)
	assert.Equal(t, len(res.Log.Records()), 2)
	assert.Equal(t, res.TotalCount, int64(2))
	assert.Equal(t, res.PaginationResult.NextPageToken, "")
}

func TestSessionLogRecordNextPageTokenNotEmpty(t *testing.T) {
	f, p, b := preSessionTest(t)

	s := f.newSession(sdktypes.SessionStateTypeCompleted, p, b)
	l := sdktypes.NewPrintSessionLogRecord("meow")
	logr, err := toSessionLogRecord(s.SessionID, l)
	assert.NoError(t, err)

	f.createSessionsAndAssert(t, s) // will create session and session record as well
	assert.NoError(t, f.gormdb.addSessionLogRecord(f.ctx, logr, ""))
	assert.NoError(t, f.gormdb.addSessionLogRecord(f.ctx, logr, ""))

	sid := sdktypes.NewIDFromUUID[sdktypes.SessionID](s.SessionID)
	res, err := f.gormdb.GetSessionLog(context.Background(),
		sdkservices.ListSessionLogRecordsFilter{
			SessionID:         sid,
			PaginationRequest: sdktypes.PaginationRequest{PageSize: 2, Ascending: true},
		})

	assert.NoError(t, err)
	assert.Equal(t, len(res.Log.Records()), 2)
	assert.Equal(t, res.TotalCount, int64(3))
	assert.NotEmpty(t, res.NextPageToken)

	// Get Next Batch to exhaust next page token
	res, err = f.gormdb.GetSessionLog(context.Background(),
		sdkservices.ListSessionLogRecordsFilter{
			SessionID:         sid,
			PaginationRequest: sdktypes.PaginationRequest{PageToken: res.PaginationResult.NextPageToken},
		})

	assert.NoError(t, err)
	assert.Equal(t, len(res.Log.Records()), 1)
	assert.Equal(t, res.TotalCount, int64(3))
	assert.Equal(t, res.PaginationResult.NextPageToken, "")
}
