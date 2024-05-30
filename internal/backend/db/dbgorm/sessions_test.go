package dbgorm

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (f *dbFixture) createSessionsAndAssert(t *testing.T, sessions ...scheme.Session) {
	for _, session := range sessions {
		assert.NoError(t, f.gormdb.createSession(f.ctx, &session))
		findAndAssertOne(t, f, session, "session_id = ?", session.SessionID)
	}
}

func (f *dbFixture) addSessionLogRecordAndAssert(t *testing.T, logr scheme.SessionLogRecord, expected int) {
	assert.NoError(t, addSessionLogRecordDB(f.gormdb.db, &logr))
	findAndAssertCount[scheme.SessionLogRecord](t, f, expected, "session_id = ?", logr.SessionID)
}

func (f *dbFixture) listSessionsAndAssert(t *testing.T, expected int) []scheme.Session {
	flt := sdkservices.ListSessionsFilter{
		CountOnly: false,
		StateType: sdktypes.SessionStateTypeUnspecified, // fetch all sesssions
	}

	sessions, cnt, err := f.gormdb.listSessions(f.ctx, flt)
	require.NoError(t, err)
	assert.Equal(t, expected, cnt)
	require.Equal(t, expected, len(sessions))

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

func TestCreateSession(t *testing.T) {
	f := preSessionTest(t)

	s := f.newSession(sdktypes.SessionStateTypeCompleted)
	// test createSession without any assets session depends on, since they are soft-foreign keys and could be nil
	f.createSessionsAndAssert(t, s)

	logs := findAndAssertCount[scheme.SessionLogRecord](t, f, 1, "session_id = ?", s.SessionID)
	assert.Equal(t, s.SessionID, logs[0].SessionID) // compare only ids, since actual log isn't empty
}

func TestCreateSessionForeignKeys(t *testing.T) {
	// check session creation if foreign keys are not nil
	f := preSessionTest(t)

	// negative test with non-existing assets
	s := f.newSession(sdktypes.SessionStateTypeCompleted)
	unexisting := uuid.New()

	s.BuildID = &unexisting
	assert.ErrorIs(t, f.gormdb.createSession(f.ctx, &s), gorm.ErrForeignKeyViolated)
	s.BuildID = nil

	s.EnvID = &unexisting
	assert.ErrorIs(t, f.gormdb.createSession(f.ctx, &s), gorm.ErrForeignKeyViolated)
	s.EnvID = nil

	s.DeploymentID = &unexisting
	assert.ErrorIs(t, f.gormdb.createSession(f.ctx, &s), gorm.ErrForeignKeyViolated)
	s.DeploymentID = nil

	s.EventID = &unexisting
	assert.ErrorIs(t, f.gormdb.createSession(f.ctx, &s), gorm.ErrForeignKeyViolated)
	s.EventID = nil

	// test with existing assets
	b := f.newBuild()
	env := f.newEnv()
	d := f.newDeployment()
	ev := f.newEvent()

	f.saveBuildsAndAssert(t, b)
	f.WithForeignKeysDisabled(func() { f.createEnvsAndAssert(t, env) })
	f.createDeploymentsAndAssert(t, d)
	f.WithForeignKeysDisabled(func() { f.createEventsAndAssert(t, ev) })

	s.BuildID = &b.BuildID
	s.EnvID = &env.EnvID
	s.DeploymentID = &d.DeploymentID
	s.EventID = &ev.EventID
	f.createSessionsAndAssert(t, s)
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

func TestCreateSessionLogRecordForeignKeys(t *testing.T) {
	f := preSessionTest(t)

	s := f.newSession(sdktypes.SessionStateTypeCompleted)
	logr := f.newSessionLogRecord()
	assert.ErrorIs(t, addSessionLogRecordDB(f.gormdb.db, &logr), gorm.ErrForeignKeyViolated)

	f.createSessionsAndAssert(t, s) // will create session and session record as well

	// test createSessionLogRecord
	f.addSessionLogRecordAndAssert(t, logr, 2) // one log record was already created due to session creation
}
