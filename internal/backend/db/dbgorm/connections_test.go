package dbgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (f *dbFixture) createConnectionsAndAssert(t *testing.T, connections ...scheme.Connection) {
	for _, conn := range connections {
		assert.NoError(t, f.gormdb.createConnection(f.ctx, &conn))
		findAndAssertOne(t, f, conn, "connection_id = ?", conn.ConnectionID)
	}
}

func (f *dbFixture) assertConnectionDeleted(t *testing.T, connections ...scheme.Connection) {
	for _, connection := range connections {
		assertDeleted(t, f, connection)
	}
}

func preConnectionTest(t *testing.T) *dbFixture {
	f := newDBFixture()
	findAndAssertCount(t, f, scheme.Connection{}, 0, "") // no connections
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

	// negative test with non-existing assets
	c := f.newConnection()

	// FIXME: ENG-571 - integration table
	// c.IntegrationID = &unexisting
	//assert.ErrorIs(t, f.gormdb.createConnection(f.ctx, &c), gorm.ErrForeignKeyViolated)
	// c.IntegrationID = nil

	c.ProjectID = scheme.UUIDOrNil(sdktypes.NewProjectID().UUIDValue())
	assert.ErrorIs(t, f.gormdb.createConnection(f.ctx, &c), gorm.ErrForeignKeyViolated)
	c.ProjectID = nil

	// test with existing assets
	p := f.newProject()
	i := f.newIntegration()
	f.createProjectsAndAssert(t, p)
	f.createIntegrationsAndAssert(t, i)

	c.IntegrationID = &i.IntegrationID
	c.ProjectID = &p.ProjectID
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
