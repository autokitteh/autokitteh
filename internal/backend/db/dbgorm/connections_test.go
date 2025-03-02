package dbgorm

import (
	"context"
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

func preConnectionTest(t *testing.T) (*dbFixture, scheme.Project) {
	f := newDBFixture()
	findAndAssertCount[scheme.Connection](t, f, 0, "") // no connections

	p := f.newProject()
	f.createProjectsAndAssert(t, p)

	return f, p
}

func TestCreateConnection(t *testing.T) {
	// test createConnection without any dependencies, since they are soft-foreign keys and could be nil
	f, p := preConnectionTest(t)

	c := f.newConnection(p)

	// test createConnection
	f.createConnectionsAndAssert(t, c)
}

func TestCreateConnectionForeignKeys(t *testing.T) {
	// test createConnection if foreign keys are not nil
	f, p := preConnectionTest(t)

	b := f.newBuild(p)
	f.saveBuildsAndAssert(t, b)

	// negative test with non-existing assets
	// use buildID as unexisting IDs and to allow us to pass ownership checks
	c := f.newConnection(p)

	// FIXME: ENG-571 - integration table
	// c.IntegrationID = &unexisting
	// assert.ErrorIs(t, f.gormdb.createConnection(f.ctx, &c), gorm.ErrForeignKeyViolated)
	// c.IntegrationID = nil

	c.ProjectID = b.BuildID // no such projectID, since it's a buildID
	assert.ErrorIs(t, f.gormdb.createConnection(f.ctx, &c), gorm.ErrForeignKeyViolated)

	// test with existing assets
	c = f.newConnection(p)
	f.createConnectionsAndAssert(t, c)
}

func TestCreateConnectionSameName(t *testing.T) {
	// test createConnection without any dependencies, since they are soft-foreign keys and could be nil
	f, p1 := preConnectionTest(t)

	// test createConnection with the same name
	connName := "same_name"

	// should prevent creating same name
	c1 := f.newConnection(connName, p1)
	c2 := f.newConnection(connName, p1)
	f.createConnectionsAndAssert(t, c1)
	assert.ErrorIs(t, f.gormdb.createConnection(f.ctx, &c2), gorm.ErrDuplicatedKey)

	// after deletion, we can create new connection with the same name in the project
	assert.NoError(t, f.gormdb.deleteConnection(f.ctx, c1.ConnectionID))
	f.createConnectionsAndAssert(t, c2)

	// and we could create connection with the same name for another project
	p2 := f.newProject()
	f.createProjectsAndAssert(t, p2)
	c5 := f.newConnection(p2, connName)
	f.createConnectionsAndAssert(t, c5)
}

func TestDeleteConnection(t *testing.T) {
	f, p := preConnectionTest(t)

	c := f.newConnection(p)
	f.createConnectionsAndAssert(t, c)

	// test deleteConnection
	assert.NoError(t, f.gormdb.deleteConnection(f.ctx, c.ConnectionID))
	f.assertConnectionDeleted(t, c)
}

func TestDeleteConnectionForeignKeys(t *testing.T) {
	f, p := preConnectionTest(t)

	c := f.newConnection(p)
	evt := f.newEvent(c, p)
	f.createConnectionsAndAssert(t, c)
	f.createEventsAndAssert(t, evt)
	// also trigger and signal are dependant on connection

	// test deleteConnection
	assert.NoError(t, f.gormdb.deleteConnection(f.ctx, c.ConnectionID))
	f.assertConnectionDeleted(t, c) // soft deleted, dependent object won't complain
}

func TestDeleteConnectionAndVars(t *testing.T) {
	ctx := context.Background()

	f, p := preConnectionTest(t)

	c1, c2, c3 := f.newConnection(p), f.newConnection(p), f.newConnection(p)
	f.createConnectionsAndAssert(t, c1, c2, c3)

	// delete specific connectionID
	assert.NoError(t, f.gormdb.deleteConnectionsAndVars(ctx, "connection_id", c1.ConnectionID))
	f.assertConnectionDeleted(t, c1)

	// delete all connections for specific projectID
	assert.NoError(t, f.gormdb.deleteConnectionsAndVars(ctx, "project_id", p.ProjectID))
	f.assertConnectionDeleted(t, c2, c3)
}

func TestCantDeleteConnectionWithTriggers(t *testing.T) {
	ctx := context.Background()
	f, p := preConnectionTest(t)

	c := f.newConnection(p)
	f.createConnectionsAndAssert(t, c)

	trigger := f.newTrigger(c, p)
	f.createTriggersAndAssert(t, trigger)

	// Try deleting a connection that has associated triggers (should fail)
	err := f.gormdb.deleteConnectionsAndVars(ctx, "connection_id", c.ConnectionID)
	assert.Error(t, err)
	assert.EqualError(t, err, "cannot delete a connection that has associated triggers")
}

func TestGetConnection(t *testing.T) {
	f, p := preConnectionTest(t)

	c := f.newConnection(p)
	f.createConnectionsAndAssert(t, c)

	// test getConnection
	c2, err := f.gormdb.getConnection(f.ctx, c.ConnectionID)

	if assert.NotEmpty(t, c2.CreatedAt) {
		assert.Equal(t, c2.CreatedAt, c2.UpdatedAt)
	}

	resetTimes(c2, &c)

	assert.NoError(t, err)
	assert.Equal(t, c, *c2)

	// test getConnection after delete
	assert.NoError(t, f.gormdb.deleteConnection(f.ctx, c.ConnectionID))
	_, err = f.gormdb.getConnection(f.ctx, c.ConnectionID)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestListConnection(t *testing.T) {
	f, p := preConnectionTest(t)

	c := f.newConnection(p)
	f.createConnectionsAndAssert(t, c)

	// test listConnection
	cc, err := f.gormdb.listConnections(f.ctx, sdkservices.ListConnectionsFilter{}, false)
	resetTimes(&cc[0])
	assert.NoError(t, err)
	assert.Len(t, cc, 1)
	assert.Equal(t, c, cc[0])

	// test listConnection after delete
	assert.NoError(t, f.gormdb.deleteConnection(f.ctx, c.ConnectionID))
	cc, err = f.gormdb.listConnections(f.ctx, sdkservices.ListConnectionsFilter{}, false)
	assert.NoError(t, err)
	assert.Len(t, cc, 0)
}

func TestDoesConnectionHaveTriggers(t *testing.T) {
	f, p := preConnectionTest(t)

	c := f.newConnection(p)
	f.createConnectionsAndAssert(t, c)

	// test doesConnectionHaveTriggers with no triggers
	hasTriggers, err := f.gormdb.doesConnectionHaveTriggers(f.ctx, c.ConnectionID)
	assert.NoError(t, err)
	assert.False(t, hasTriggers)

	trigger := f.newTrigger(c, p)
	f.createTriggersAndAssert(t, trigger)

	// test doesConnectionHaveTriggers with triggers
	hasTriggers, err = f.gormdb.doesConnectionHaveTriggers(f.ctx, c.ConnectionID)
	assert.NoError(t, err)
	assert.True(t, hasTriggers)

	assert.NoError(t, f.gormdb.deleteTrigger(f.ctx, trigger.TriggerID))

	// test doesConnectionHaveTriggers after trigger delete
	hasTriggers, err = f.gormdb.doesConnectionHaveTriggers(f.ctx, c.ConnectionID)
	assert.NoError(t, err)
	assert.False(t, hasTriggers)
}
