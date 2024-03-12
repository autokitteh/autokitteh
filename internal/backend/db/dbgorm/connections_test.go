package dbgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
)

func (f *dbFixture) createConnectionsAndAssert(t *testing.T, connections ...scheme.Connection) {
	for _, conn := range connections {
		assert.NoError(t, f.gormdb.createConnection(f.ctx, &conn))
		findAndAssertOne(t, f, conn, "connection_id = ?", conn.ConnectionID)
	}
}

func TestCreateConnection(t *testing.T) {
	f := newDBFixture(true)                              // no foreign keys
	findAndAssertCount(t, f, scheme.Connection{}, 0, "") // no connections

	tr := f.newConnection()
	// test createConnection
	f.createConnectionsAndAssert(t, tr)
}
