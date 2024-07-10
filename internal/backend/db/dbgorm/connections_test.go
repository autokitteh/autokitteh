package dbgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
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
	i := f.newIntegration("test")
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

func TestDeleteConnectionAndVars(t *testing.T) {
	f := preConnectionTest(t).WithDebug()

	p := f.newProject()
	c1, c2, c3 := f.newConnection(p), f.newConnection(p), f.newConnection(p)
	f.createProjectsAndAssert(t, p)
	f.createConnectionsAndAssert(t, c1, c2, c3)

	// delete specific connectionID
	assert.NoError(t, f.gormdb.deleteConnectionsAndVars("connection_id", c1.ConnectionID))
	f.assertConnectionDeleted(t, c1)

	// delete all connections for specific projectID
	assert.NoError(t, f.gormdb.deleteConnectionsAndVars("project_id", p.ProjectID))
	f.assertConnectionDeleted(t, c2, c3)
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

func TestListConnection(t *testing.T) {
	f := preConnectionTest(t)

	c := f.newConnection()
	f.createConnectionsAndAssert(t, c)

	// test listConnection
	cc, err := f.gormdb.listConnections(f.ctx, sdkservices.ListConnectionsFilter{}, false)
	assert.NoError(t, err)
	assert.Len(t, cc, 1)
	assert.Equal(t, c, cc[0])

	// test listConnection after delete
	assert.NoError(t, f.gormdb.deleteConnection(f.ctx, c.ConnectionID))
	cc, err = f.gormdb.listConnections(f.ctx, sdkservices.ListConnectionsFilter{}, false)
	assert.NoError(t, err)
	assert.Len(t, cc, 0)
}
