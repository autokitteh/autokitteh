package dbgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
)

func (f *dbFixture) createConnectionsAndAssert(t *testing.T, connections ...scheme.Connection) {
	for _, conn := range connections {
		assert.NoError(t, f.gormdb.createConnection(f.ctx, &conn))
		findAndAssertOne(t, f, conn, "connection_id = ?", conn.ConnectionID)
	}
}

func (f *dbFixture) assertConnectionDeleted(t *testing.T, connections ...scheme.Connection) {
	for _, connection := range connections {
		assertSoftDeleted(t, f, connection)
	}
}

func preConnectionTest(t *testing.T) *dbFixture {
	f := newDBFixture()
	findAndAssertCount[scheme.Connection](t, f, 0, "") // no connections
	return f
}

func TestCreateConnection(t *testing.T) {
	// test createConneciton without any dependencies, since they are soft-foreign keys and could be nil
	f := preConnectionTest(t)

	c := f.newConnection()

	// test createConnection
	f.createConnectionsAndAssert(t, c)
}

func TestCreateConnectionForeignKeys(t *testing.T) {
	// test createConnection if foreign keys are not nil
	f := preConnectionTest(t)

	p := f.newProject()
	i := f.newIntegration()
	b := f.newBuild()
	f.saveBuildsAndAssert(t, b)
	f.createProjectsAndAssert(t, p)
	f.createIntegrationsAndAssert(t, i)

	// negative test with non-existing assets
	// use buildID as unexisting IDs and to allow us to pass ownership checks
	c := f.newConnection()

	// FIXME: ENG-571 - integration table
	// c.IntegrationID = &unexisting
	// assert.ErrorIs(t, f.gormdb.createConnection(f.ctx, &c), gorm.ErrForeignKeyViolated)
	// c.IntegrationID = nil

	c.ProjectID = &b.BuildID // no such projectID, since it's a buildID
	assert.ErrorIs(t, f.gormdb.createConnection(f.ctx, &c), gorm.ErrForeignKeyViolated)

	// test with existing assets
	c = f.newConnection(i, p, i)
	f.createConnectionsAndAssert(t, c)
}

func TestDeleteConnection(t *testing.T) {
	f := preConnectionTest(t)

	c := f.newConnection()
	f.createConnectionsAndAssert(t, c)

	// test deleteConnection
	assert.NoError(t, f.gormdb.deleteConnection(f.ctx, c.ConnectionID))
	f.assertConnectionDeleted(t, c)
}

func TestDeleteConnectionForeignKeys(t *testing.T) {
	f := preConnectionTest(t)

	c := f.newConnection()
	evt := f.newEvent(c)
	f.createConnectionsAndAssert(t, c)
	f.createEventsAndAssert(t, evt)
	// also trigger and signal are dependant on connection

	// test deleteConnection
	assert.NoError(t, f.gormdb.deleteConnection(f.ctx, c.ConnectionID))
	f.assertConnectionDeleted(t, c) // soft deleted, dependent object won't complain
}

func TestGetConnection(t *testing.T) {
	f := preConnectionTest(t)

	c := f.newConnection()
	f.createConnectionsAndAssert(t, c)

	// test getConnection
	c2, err := f.gormdb.getConnection(f.ctx, c.ConnectionID)
	assert.NoError(t, err)
	assert.Equal(t, c, *c2)

	// test getConnection after delete
	assert.NoError(t, f.gormdb.deleteConnection(f.ctx, c.ConnectionID))
	_, err = f.gormdb.getConnection(f.ctx, c.ConnectionID)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
