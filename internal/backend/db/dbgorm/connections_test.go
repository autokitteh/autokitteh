package dbgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
)

func createConnectionsAndAssert(t *testing.T, f *dbFixture, connections ...scheme.Connection) {
	for _, conn := range connections {
		assert.NoError(t, f.gormdb.createConnection(f.ctx, &conn))
		findAndAssertOne(t, f, conn, "connection_id = ?", conn.ConnectionID)
	}
}

func TestCreateConnection(t *testing.T) {
	f := newDBFixture(true)                              // no foreign keys
	findAndAssertCount(t, f, scheme.Connection{}, 0, "") // no connections

	tr := newConnection()
	// test createConnection
	createConnectionsAndAssert(t, f, tr)
}
