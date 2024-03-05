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

func createSessionAndAssert(t *testing.T, f *dbFixture, session scheme.Session) {
	assert.NoError(t, f.gormdb.createSession(f.ctx, session))
	findAndAssertOne(t, f, session, "session_id = ?", session.SessionID)
}

func listSessionsAndAssert(t *testing.T, f *dbFixture, expected int) []scheme.Session {
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

// check session status directly via GORM (not via our gormdb API)
func assertSessionDeleted(t *testing.T, f *dbFixture, sessionID string) {
	// check that session is not found without unscoped
	res := f.db.First(&scheme.Session{}, "session_id = ?", sessionID)
	assert.ErrorAs(t, gorm.ErrRecordNotFound, &res.Error)

	// check that session is marked as deleted
	res = f.db.Unscoped().First(&scheme.Session{}, "session_id = ?", sessionID)
	assert.NoError(t, res.Error)
	assert.Equal(t, int64(1), res.RowsAffected)
	var session scheme.Session
	res.Scan(&session)
	assert.NotNil(t, session.DeletedAt)
}

func TestCreateSession(t *testing.T) {
	f := newDbFixture(true)        // no foreign keys
	listSessionsAndAssert(t, f, 0) // no sessions

	s := newSession(f, sdktypes.SessionStateTypeCompleted)
	// test createSession
	createSessionAndAssert(t, f, s)

	logs := findAndAssertCount(t, f, scheme.SessionLogRecord{}, 1, "session_id = ?", s.SessionID)
	assert.Equal(t, s.SessionID, logs[0].SessionID) // compare only ids, since actual log isn't empty
}

func TestGetSession(t *testing.T) {
	f := newDbFixture(true)        // no foreign keys
	listSessionsAndAssert(t, f, 0) // no sessions

	s := newSession(f, sdktypes.SessionStateTypeCompleted)
	createSessionAndAssert(t, f, s)

	// check getSession
	session, err := f.gormdb.getSession(f.ctx, s.SessionID)
	assert.NoError(t, err)
	assert.Equal(t, s, *session)

	// check that after deleteSession it's not found
	assert.NoError(t, f.gormdb.deleteSession(f.ctx, s.SessionID))
	_, err = f.gormdb.getSession(f.ctx, s.SessionID)
	assert.ErrorAs(t, err, &gorm.ErrRecordNotFound)
}

func TestListSessions(t *testing.T) {
	f := newDbFixture(true)        // no foreign keys
	listSessionsAndAssert(t, f, 0) // no sessions

	s := newSession(f, sdktypes.SessionStateTypeCompleted)
	createSessionAndAssert(t, f, s)

	sessions := listSessionsAndAssert(t, f, 1)
	assert.Equal(t, s, sessions[0])

	// deleteSession and ensure that listSessions is empty
	assert.NoError(t, f.gormdb.deleteSession(f.ctx, s.SessionID))
	listSessionsAndAssert(t, f, 0)
}

func TestDeleteSession(t *testing.T) {
	f := newDbFixture(true)        // no foreign keys
	listSessionsAndAssert(t, f, 0) // no sessions

	s := newSession(f, sdktypes.SessionStateTypeCompleted)
	createSessionAndAssert(t, f, s)

	assert.NoError(t, f.gormdb.deleteSession(f.ctx, s.SessionID))
	listSessionsAndAssert(t, f, 0)
	assertSessionDeleted(t, f, s.SessionID)
}

func TestForeignKeysSession(t *testing.T) {
	f := newDbFixture(false)       // with foreign keys
	listSessionsAndAssert(t, f, 0) // no sessions

	s := newSession(f, sdktypes.SessionStateTypeCompleted)
	err := f.gormdb.createSession(f.ctx, s) // should fail since there is no deployment
	assert.ErrorContains(t, err, "FOREIGN KEY")
}
