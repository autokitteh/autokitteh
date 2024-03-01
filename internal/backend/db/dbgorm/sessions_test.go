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

func createSessionWithTest(t *testing.T, f *dbFixture, session scheme.Session) {
	assert.Nil(t, f.gormdb.createSession(f.ctx, session))
	res := f.gormdb.db.First(&scheme.Session{}, "session_id = ?", session.SessionID)
	require.Nil(t, res.Error)
	assert.Equal(t, int64(1), res.RowsAffected)
}

func listSessionsAndTest(t *testing.T, f *dbFixture, expected int) []scheme.Session {
	flt := sdkservices.ListSessionsFilter{
		CountOnly: false,
		StateType: sdktypes.UnspecifiedSessionStateType, // fetch all sesssions
	}

	sessions, cnt, err := f.gormdb.listSessions(f.ctx, flt)
	require.Nil(t, err)
	assert.Equal(t, expected, cnt)
	assert.Equal(t, expected, len(sessions))

	return sessions
}

// check session status directly via GORM (not via our gormdb API)
func testSessionDeleted(t *testing.T, f *dbFixture, sessionID string) {
	// check that session is not found without unscoped
	res := f.db.First(&scheme.Session{}, "session_id = ?", sessionID)
	assert.Equal(t, gorm.ErrRecordNotFound, res.Error)

	// check that session is marked as deleted
	res = f.db.Unscoped().First(&scheme.Session{}, "session_id = ?", sessionID)
	assert.Nil(t, res.Error)
	assert.Equal(t, int64(1), res.RowsAffected)
	var session scheme.Session
	res.Scan(&session)
	assert.NotNil(t, session.DeletedAt)
}

func TestCreateSession(t *testing.T) {
	f := newDbFixture(true)      // no foreign keys
	listSessionsAndTest(t, f, 0) // no sessions

	s := newSession(f, sdktypes.CompletedSessionStateType)
	createSessionWithTest(t, f, s)

	// obtain all session records from the session table
	var sessions []scheme.Session
	assert.NoError(t, f.db.Find(&sessions).Error)
	assert.Equal(t, int(1), len(sessions))
	assert.Equal(t, s, sessions[0])

	var sessionLogs []scheme.SessionLogRecord
	assert.NoError(t, f.db.Find(&sessionLogs).Error)
	assert.Equal(t, int(1), len(sessionLogs))
	assert.Equal(t, s.SessionID, sessionLogs[0].SessionID)
}

func TestGetSession(t *testing.T) {
	f := newDbFixture(true)      // no foreign keys
	listSessionsAndTest(t, f, 0) // no sessions

	s := newSession(f, sdktypes.CompletedSessionStateType)
	createSessionWithTest(t, f, s)

	// check get session
	session, err := f.gormdb.getSession(f.ctx, s.SessionID)
	assert.Nil(t, err)
	assert.Equal(t, s, *session)

	// check that after delete session is not found
	assert.Nil(t, f.gormdb.deleteSession(f.ctx, s.SessionID))
	_, err = f.gormdb.getSession(f.ctx, s.SessionID)
	assert.Equal(t, err, gorm.ErrRecordNotFound)
}

func TestListSessions(t *testing.T) {
	f := newDbFixture(true)      // no foreign keys
	listSessionsAndTest(t, f, 0) // no sessions

	s := newSession(f, sdktypes.CompletedSessionStateType)
	createSessionWithTest(t, f, s)

	sessions := listSessionsAndTest(t, f, 1)
	assert.Equal(t, s, sessions[0])

	// delete and ensure that list is empty
	assert.Nil(t, f.gormdb.deleteSession(f.ctx, s.SessionID))
	listSessionsAndTest(t, f, 0)

	testSessionDeleted(t, f, s.SessionID)
}

func TestDeleteSession(t *testing.T) {
	f := newDbFixture(true)      // no foreign keys
	listSessionsAndTest(t, f, 0) // no sessions

	s := newSession(f, sdktypes.CompletedSessionStateType)
	createSessionWithTest(t, f, s)

	assert.Nil(t, f.gormdb.deleteSession(f.ctx, s.SessionID))
	listSessionsAndTest(t, f, 0)

	testSessionDeleted(t, f, s.SessionID)
}

func TestForeignKeysSession(t *testing.T) {
	f := newDbFixture(false)     // with foreign keys
	listSessionsAndTest(t, f, 0) // no sessions

	s := newSession(f, sdktypes.CompletedSessionStateType)
	err := f.gormdb.createSession(f.ctx, s) // should fail since there is no deployment
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "FOREIGN KEY")
}
