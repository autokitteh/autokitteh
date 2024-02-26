package dbgorm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.autokitteh.dev/autokitteh/backend/internal/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"gorm.io/gorm"
)

func createSessionWithTest(t *testing.T, ctx context.Context, gormdb *gormdb, session scheme.Session) {
	assert.Nil(t, gormdb.createSession(ctx, session))
	res := gormdb.db.First(&scheme.Session{}, "session_id = ?", session.SessionID)
	assert.Nil(t, res.Error)
	assert.Equal(t, int64(1), res.RowsAffected)
}

func TestCreateSession(t *testing.T) {
	f := newDbFixture()

	session := makeSchemeSession()
	createSessionWithTest(t, f.ctx, f.gormdb, session)

	// obtain all session records from the session table
	var sessions []scheme.Session
	assert.Nil(t, f.db.Find(&sessions).Error)
	assert.Equal(t, int(1), len(sessions))
	assert.Equal(t, session, sessions[0])

	var sessionLogs []scheme.SessionLogRecord
	assert.Nil(t, f.db.Find(&sessionLogs).Error)
	assert.Equal(t, int(1), len(sessionLogs))
	assert.Equal(t, session.SessionID, sessionLogs[0].SessionID)
}

func TestDeleteSession(t *testing.T) {
	f := newDbFixture()

	session := makeSchemeSession()
	createSessionWithTest(t, f.ctx, f.gormdb, session)

	assert.Nil(t, f.gormdb.deleteSession(f.ctx, session.SessionID))

	// check that session is ignored without unscoped
	res := f.db.First(&scheme.Session{}, "session_id = ?", session.SessionID)
	assert.Equal(t, res.Error, gorm.ErrRecordNotFound)

	// check that session is marked as deleted
	res = f.db.Unscoped().First(&scheme.Session{}, "session_id = ?", session.SessionID)
	assert.Nil(t, res.Error)
	assert.Equal(t, int64(1), res.RowsAffected)
	res.Scan(&session)
	assert.NotNil(t, session.DeletedAt)
}

func TestListSessions(t *testing.T) {
	f := newDbFixture()

	flt := sdkservices.ListSessionsFilter{
		CountOnly: false,
		StateType: sdktypes.UnspecifiedSessionStateType,
	}

	// no sessions
	sessions, cnt, err := f.gormdb.listSessions(f.ctx, flt)
	assert.Nil(t, err)
	assert.Equal(t, 0, cnt)
	assert.Equal(t, 0, len(sessions))

	// create session and obtain it via list
	session := makeSchemeSession()
	createSessionWithTest(t, f.ctx, f.gormdb, session)

	sessions, cnt, err = f.gormdb.listSessions(f.ctx, flt)
	assert.Nil(t, err)
	assert.Equal(t, 1, cnt)
	assert.Equal(t, 1, len(sessions))
	assert.Equal(t, session.SessionID, sessions[0].SessionID)

	// delete and ensure that list is empty
	assert.Nil(t, f.gormdb.deleteSession(f.ctx, session.SessionID))

	sessions, cnt, err = f.gormdb.listSessions(f.ctx, flt)
	assert.Nil(t, err)
	assert.Equal(t, 0, cnt)
	assert.Equal(t, 0, len(sessions))
}
