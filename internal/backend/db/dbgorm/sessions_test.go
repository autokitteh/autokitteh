package dbgorm

import (
	"testing"

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
		assertSoftDeleted(t, f, scheme.Session{SessionID: session.SessionID})
	}
}

func TestCreateSession(t *testing.T) {
	f := newDBFixture(true)       // no foreign keys
	f.listSessionsAndAssert(t, 0) // no sessions

	s := f.newSession(sdktypes.SessionStateTypeCompleted)
	// test createSession
	f.createSessionsAndAssert(t, s)

	logs := findAndAssertCount(t, f, scheme.SessionLogRecord{}, 1, "session_id = ?", s.SessionID)
	assert.Equal(t, s.SessionID, logs[0].SessionID) // compare only ids, since actual log isn't empty
}

func TestCreateSessionForeignKeys(t *testing.T) {
	f := newDBFixture(false)      // with foreign keys
	f.listSessionsAndAssert(t, 0) // no sessions

	s := f.newSession(sdktypes.SessionStateTypeCompleted)
	err := f.gormdb.createSession(f.ctx, &s) // should fail since there is no deployment
	assert.ErrorContains(t, err, "FOREIGN KEY")
}

func TestGetSession(t *testing.T) {
	f := newDBFixture(true)       // no foreign keys
	f.listSessionsAndAssert(t, 0) // no sessions

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
	f := newDBFixture(true)       // no foreign keys
	f.listSessionsAndAssert(t, 0) // no sessions

	s := f.newSession(sdktypes.SessionStateTypeCompleted)
	f.createSessionsAndAssert(t, s)

	sessions := f.listSessionsAndAssert(t, 1)
	assert.Equal(t, s, sessions[0])

	// deleteSession and ensure that listSessions is empty
	assert.NoError(t, f.gormdb.deleteSession(f.ctx, s.SessionID))
	f.listSessionsAndAssert(t, 0)
}

func TestDeleteSession(t *testing.T) {
	f := newDBFixture(true)       // no foreign keys
	f.listSessionsAndAssert(t, 0) // no sessions

	s := f.newSession(sdktypes.SessionStateTypeCompleted)
	f.createSessionsAndAssert(t, s)

	assert.NoError(t, f.gormdb.deleteSession(f.ctx, s.SessionID))
	f.assertSessionsDeleted(t, s)
}
