package dbgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (f *dbFixture) createSessionsAndAssert(t *testing.T, sessions ...scheme.Session) {
	for _, session := range sessions {
		assert.NoError(t, f.gormdb.createSession(f.ctx, &session))
		findAndAssertOne(t, f, session, "session_id = ?", session.SessionID)
	}
}

func (f *dbFixture) listSessionsAndAssert(t *testing.T, expected int64) []scheme.Session {
	flt := sdkservices.ListSessionsFilter{
		CountOnly: false,
		StateType: sdktypes.SessionStateTypeUnspecified, // fetch all sesssions
	}

	sessions, cnt, err := f.gormdb.listSessions(f.ctx, flt)
	require.NoError(t, err)
	assert.Equal(t, expected, cnt)
	require.Equal(t, expected, int64(len(sessions)))

	return sessions
}

func (f *dbFixture) assertSessionsDeleted(t *testing.T, sessions ...scheme.Session) {
	for _, session := range sessions {
		assertSoftDeleted(t, f, session)
	}
}

func preSessionTest(t *testing.T) *dbFixture {
	f := newDBFixture()
	f.listSessionsAndAssert(t, 0) // no sessions
	findAndAssertCount[scheme.SessionLogRecord](t, f, 0, "")
	return f
}

func testLastLogRecord(t *testing.T, f *dbFixture, numRecords int, sessionID sdktypes.UUID, lr sdktypes.SessionLogRecord) {
	logs, err := f.gormdb.getSessionLogRecords(f.ctx, sessionID)
	assert.NoError(t, err)
	assert.Equal(t, numRecords, len(logs))
	for _, r := range logs {
		assert.Equal(t, sessionID, r.SessionID)
	}

	l, err := scheme.ParseSessionLogRecord(logs[len(logs)-1]) // last log
	assert.NoError(t, err)

	l = l.WithoutTimestamp()
	lr = lr.WithProcessID(fixtures.ProcessID()).WithoutTimestamp()
	assert.Equal(t, l, lr)
}

func TestCreateSession(t *testing.T) {
	f := preSessionTest(t)

	// test createSession
	s := f.newSession(sdktypes.SessionStateTypeCompleted)
	f.createSessionsAndAssert(t, s) // all session assets are optional and will be set to nil

	// test getSessionLogRecords and ensure that session logs contain the only CREATED record
	testLastLogRecord(t, f, 1, s.SessionID, sdktypes.NewStateSessionLogRecord(sdktypes.NewSessionStateCreated()))
}

func TestCreateSessionForeignKeys(t *testing.T) {
	// check session creation if foreign keys are not nil
	f := preSessionTest(t)

	// test with existing assets
	p := f.newProject()
	b := f.newBuild()
	env := f.newEnv(p)
	d := f.newDeployment(b)
	evt := f.newEvent()
	s := f.newSession(sdktypes.SessionStateTypeCompleted, d, b, env, evt)

	f.createProjectsAndAssert(t, p)
	f.saveBuildsAndAssert(t, b)
	f.createEnvsAndAssert(t, env)
	f.createDeploymentsAndAssert(t, d)
	f.createEventsAndAssert(t, evt)
	f.createSessionsAndAssert(t, s)

	// negative test with non-existing assets
	// use existing user owned projectID as fakeID to pass user check

	s2 := f.newSession(sdktypes.SessionStateTypeCompleted)

	s2.BuildID = &p.ProjectID // no such buildID, since it's a projectID
	assert.ErrorIs(t, f.gormdb.createSession(f.ctx, &s2), gorm.ErrForeignKeyViolated)
	s2.BuildID = nil

	s2.EnvID = &p.ProjectID // no such envID, since it's a projectID
	assert.ErrorIs(t, f.gormdb.createSession(f.ctx, &s2), gorm.ErrForeignKeyViolated)
	s2.EnvID = nil

	s2.DeploymentID = &p.ProjectID // no such deploymentID, since it's a projectID
	assert.ErrorIs(t, f.gormdb.createSession(f.ctx, &s2), gorm.ErrForeignKeyViolated)
	s2.DeploymentID = nil

	s2.EventID = &p.ProjectID // no such eventID, since it's a projectID
	assert.ErrorIs(t, f.gormdb.createSession(f.ctx, &s2), gorm.ErrForeignKeyViolated)
	s2.EventID = nil
}

func TestGetSession(t *testing.T) {
	f := preSessionTest(t)

	s := f.newSession(sdktypes.SessionStateTypeCompleted)
	f.createSessionsAndAssert(t, s)

	// check getSession
	session, err := f.gormdb.getSession(f.ctx, s.SessionID)
	assert.NoError(t, err)
	assert.Equal(t, s, *session)

	// check that after deleteSession it's not found
	assert.NoError(t, f.gormdb.deleteSession(f.ctx, s.SessionID))
	_, err = f.gormdb.getSession(f.ctx, s.SessionID)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestListSessions(t *testing.T) {
	f := preSessionTest(t)

	s := f.newSession(sdktypes.SessionStateTypeCompleted)
	f.createSessionsAndAssert(t, s)

	sessions := f.listSessionsAndAssert(t, 1)
	s.Inputs = nil
	assert.Equal(t, s, sessions[0])

	// deleteSession and ensure that listSessions is empty
	assert.NoError(t, f.gormdb.deleteSession(f.ctx, s.SessionID))
	f.listSessionsAndAssert(t, 0)
}

func TestListSessionsNoSessions(t *testing.T) {
	f := preSessionTest(t)

	sessions := f.listSessionsAndAssert(t, 0)
	assert.Equal(t, sessions, []scheme.Session{})
}

func TestListSessionsCountOnly(t *testing.T) {
	f := preSessionTest(t)

	s := f.newSession(sdktypes.SessionStateTypeCompleted)
	f.createSessionsAndAssert(t, s)

	flt := sdkservices.ListSessionsFilter{
		CountOnly: true,
		StateType: sdktypes.SessionStateTypeUnspecified, // fetch all sesssions
	}

	sessions, cnt, err := f.gormdb.listSessions(f.ctx, flt)
	require.NoError(t, err)
	require.Equal(t, cnt, int64(1))
	require.Nil(t, sessions)
}

func TestListSessionsNoSessionsCountOnly(t *testing.T) {
	f := preSessionTest(t)

	flt := sdkservices.ListSessionsFilter{
		CountOnly: true,
		StateType: sdktypes.SessionStateTypeUnspecified, // fetch all sesssions
	}

	sessions, cnt, err := f.gormdb.listSessions(f.ctx, flt)
	require.NoError(t, err)
	require.Equal(t, cnt, int64(0))
	require.Nil(t, sessions)
}

func TestListPaginatedSession(t *testing.T) {
	f := preSessionTest(t)

	s := f.newSession(sdktypes.SessionStateTypeCompleted)
	f.createSessionsAndAssert(t, s)

	s = f.newSession(sdktypes.SessionStateTypeCompleted)
	f.createSessionsAndAssert(t, s)

	flt := sdkservices.ListSessionsFilter{
		CountOnly:         false,
		StateType:         sdktypes.SessionStateTypeUnspecified, // fetch all sesssions
		PaginationRequest: sdktypes.PaginationRequest{PageSize: 1},
	}

	sessions, cnt, err := f.gormdb.listSessions(f.ctx, flt)
	require.NoError(t, err)
	require.Equal(t, cnt, int64(2))
	require.Len(t, sessions, 1)
}

func TestDeleteSession(t *testing.T) {
	f := preSessionTest(t)

	s := f.newSession(sdktypes.SessionStateTypeCompleted)
	f.createSessionsAndAssert(t, s)

	assert.NoError(t, f.gormdb.deleteSession(f.ctx, s.SessionID))
	f.assertSessionsDeleted(t, s)
}

/*
func TestDeleteSessionForeignKeys(t *testing.T) {
    // session is soft-deleted, so no need to check foreign keys meanwhile
}
*/

func TestAddSessionLogRecordUnexistingSession(t *testing.T) {
	f := preSessionTest(t)

	s := f.newSession(sdktypes.SessionStateTypeCompleted)
	logr := f.newSessionLogRecord() // invalid sessionID

	// NOTE: although sessionID is foreignKey for sessionLogRecord, we won't fail with gorm.ErrForeignKeyViolated
	// due to the users, we will fail with sdkerrors.ErrUnauthorized, since there is no such session for the user
	assert.ErrorIs(t, f.gormdb.addSessionLogRecord(f.ctx, &logr), sdkerrors.ErrUnauthorized)

	f.createSessionsAndAssert(t, s) // will create session and session record as well

	// ensure that with valid sessionID log could be added
	logr.SessionID = s.SessionID
	assert.NoError(t, f.gormdb.addSessionLogRecord(f.ctx, &logr))
}

func TestAddSessionPrintLogRecord(t *testing.T) {
	f := preSessionTest(t)

	s := f.newSession(sdktypes.SessionStateTypeCompleted)
	l := sdktypes.NewPrintSessionLogRecord("meow")
	logr, err := toSessionLogRecord(s.SessionID, l)
	assert.NoError(t, err)

	f.createSessionsAndAssert(t, s) // will create session and session record as well
	assert.NoError(t, f.gormdb.addSessionLogRecord(f.ctx, logr))

	testLastLogRecord(t, f, 2, s.SessionID, l)
}
