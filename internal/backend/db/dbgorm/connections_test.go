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

func (f *dbFixture) createConnectionsAndAssert(t *testing.T, connections ...scheme.Connection) {
	for _, conn := range connections {
		require.NoError(t, f.gormdb.createConnection(f.ctx, &conn))
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

	c.ProjectID = &b.BuildID // no such projectID, since it's a buildID
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
	ctx := t.Context()

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
	ctx := t.Context()
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

// TestCreateOrgLevelConnection tests creating a connection at the org level (without project_id)
func TestCreateOrgLevelConnection(t *testing.T) {
	f := newDBFixture()
	findAndAssertCount[scheme.Connection](t, f, 0, "") // no connections

	orgID := f.newOrg()

	// Create an org-level connection (no project_id)
	c := f.newConnection()
	c.ProjectID = nil // explicitly set to nil for org-level connection
	c.OrgID = orgID
	c.Name = "org_level_connection"

	f.createConnectionsAndAssert(t, c)

	// Verify the connection was created at org level
	retrieved, err := f.gormdb.getConnection(f.ctx, c.ConnectionID)
	assert.NoError(t, err)
	assert.Nil(t, retrieved.ProjectID, "org-level connection should have nil project_id")
	assert.Equal(t, orgID, retrieved.OrgID)
	assert.Equal(t, "org_level_connection", retrieved.Name)
}

// TestCreateProjectLevelConnection tests creating a connection at the project level
func TestCreateProjectLevelConnection(t *testing.T) {
	f, p := preConnectionTest(t)

	// Create a project-level connection
	c := f.newConnection(p)
	c.Name = "project_level_connection"

	f.createConnectionsAndAssert(t, c)

	// Verify the connection was created at project level
	retrieved, err := f.gormdb.getConnection(f.ctx, c.ConnectionID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved.ProjectID, "project-level connection should have a project_id")
	assert.Equal(t, p.ProjectID, *retrieved.ProjectID)
	assert.Equal(t, p.OrgID, retrieved.OrgID)
	assert.Equal(t, "project_level_connection", retrieved.Name)
}

// TestOrgLevelAndProjectLevelConnectionsWithSameName tests that org-level and project-level
// connections can have the same name within the same org
func TestOrgLevelAndProjectLevelConnectionsWithSameName(t *testing.T) {
	f := newDBFixture()
	orgID := f.newOrg()

	// Create a project in the org
	p := f.newProject(orgID)
	f.createProjectsAndAssert(t, p)

	connName := "shared_connection_name"

	// Create an org-level connection
	orgConn := f.newConnection()
	orgConn.ProjectID = nil
	orgConn.OrgID = orgID
	orgConn.Name = connName
	f.createConnectionsAndAssert(t, orgConn)

	// Create a project-level connection with the same name - should succeed
	projectConn := f.newConnection(p)
	projectConn.Name = connName
	f.createConnectionsAndAssert(t, projectConn)

	// Verify both exist
	orgRetrieved, err := f.gormdb.getConnection(f.ctx, orgConn.ConnectionID)
	assert.NoError(t, err)
	assert.Nil(t, orgRetrieved.ProjectID)
	assert.Equal(t, connName, orgRetrieved.Name)

	projectRetrieved, err := f.gormdb.getConnection(f.ctx, projectConn.ConnectionID)
	assert.NoError(t, err)
	assert.NotNil(t, projectRetrieved.ProjectID)
	assert.Equal(t, connName, projectRetrieved.Name)
}

// TestUniqueConstraintOrgLevelConnections tests that org-level connections
// must have unique names within an org
func TestUniqueConstraintOrgLevelConnections(t *testing.T) {
	f := newDBFixture()
	orgID := f.newOrg()

	connName := "unique_org_connection"

	// Create first org-level connection
	c1 := f.newConnection()
	c1.ProjectID = nil
	c1.OrgID = orgID
	c1.Name = connName
	f.createConnectionsAndAssert(t, c1)

	// Try to create another org-level connection with the same name - should fail
	c2 := f.newConnection()
	c2.ProjectID = nil
	c2.OrgID = orgID
	c2.Name = connName
	assert.ErrorIs(t, f.gormdb.createConnection(f.ctx, &c2), gorm.ErrDuplicatedKey)

	// After deletion, we can create a new connection with the same name
	assert.NoError(t, f.gormdb.deleteConnection(f.ctx, c1.ConnectionID))
	f.createConnectionsAndAssert(t, c2)
}

// TestUniqueConstraintProjectLevelConnections tests that project-level connections
// must have unique names within an org+project combination
func TestUniqueConstraintProjectLevelConnections(t *testing.T) {
	f, p := preConnectionTest(t)

	connName := "unique_project_connection"

	// Create first project-level connection
	c1 := f.newConnection(p)
	c1.Name = connName
	f.createConnectionsAndAssert(t, c1)

	// Try to create another project-level connection in the same project with the same name - should fail
	c2 := f.newConnection(p)
	c2.Name = connName
	assert.ErrorIs(t, f.gormdb.createConnection(f.ctx, &c2), gorm.ErrDuplicatedKey)

	// After deletion, we can create a new connection with the same name
	assert.NoError(t, f.gormdb.deleteConnection(f.ctx, c1.ConnectionID))
	f.createConnectionsAndAssert(t, c2)
}

// TestListConnectionsByOrg tests listing connections filtered by org
func TestListConnectionsByOrg(t *testing.T) {
	f := newDBFixture()
	org1 := f.newOrg()
	org2 := f.newOrg()

	// Create org-level connection for org1
	c1 := f.newConnection()
	c1.ProjectID = nil
	c1.OrgID = org1
	c1.Name = "org1_connection"
	f.createConnectionsAndAssert(t, c1)

	// Create project in org1 with project-level connection
	p1 := f.newProject(org1)
	f.createProjectsAndAssert(t, p1)
	c2 := f.newConnection(p1)
	c2.Name = "org1_project_connection"
	f.createConnectionsAndAssert(t, c2)

	// Create org-level connection for org2
	c3 := f.newConnection()
	c3.ProjectID = nil
	c3.OrgID = org2
	c3.Name = "org2_connection"
	f.createConnectionsAndAssert(t, c3)

	// List connections for org1
	connections, err := f.gormdb.listConnections(f.ctx, sdkservices.ListConnectionsFilter{
		OrgID: sdktypes.NewIDFromUUID[sdktypes.OrgID](org1),
	}, false)
	assert.NoError(t, err)
	assert.Len(t, connections, 2, "org1 should have 2 connections (1 org-level, 1 project-level)")

	// List connections for org2
	connections, err = f.gormdb.listConnections(f.ctx, sdkservices.ListConnectionsFilter{
		OrgID: sdktypes.NewIDFromUUID[sdktypes.OrgID](org2),
	}, false)
	assert.NoError(t, err)
	assert.Len(t, connections, 1, "org2 should have 1 connection")
}

// TestListConnectionsByProject tests listing connections filtered by project
func TestListConnectionsByProject(t *testing.T) {
	f := newDBFixture()
	org := f.newOrg()

	// Create org-level connection
	orgConn := f.newConnection()
	orgConn.ProjectID = nil
	orgConn.OrgID = org
	orgConn.Name = "org_connection"
	f.createConnectionsAndAssert(t, orgConn)

	// Create two projects
	p1 := f.newProject(org)
	f.createProjectsAndAssert(t, p1)
	p2 := f.newProject(org)
	f.createProjectsAndAssert(t, p2)

	// Create project-level connections
	c1 := f.newConnection(p1)
	c1.Name = "project1_connection"
	f.createConnectionsAndAssert(t, c1)

	c2 := f.newConnection(p2)
	c2.Name = "project2_connection"
	f.createConnectionsAndAssert(t, c2)

	// List connections for project1
	connections, err := f.gormdb.listConnections(f.ctx, sdkservices.ListConnectionsFilter{
		ProjectID: sdktypes.NewIDFromUUID[sdktypes.ProjectID](p1.ProjectID),
	}, false)
	assert.NoError(t, err)
	assert.Len(t, connections, 1, "project1 should have 1 connection")
	assert.Equal(t, "project1_connection", connections[0].Name)

	// List connections for project2
	connections, err = f.gormdb.listConnections(f.ctx, sdkservices.ListConnectionsFilter{
		ProjectID: sdktypes.NewIDFromUUID[sdktypes.ProjectID](p2.ProjectID),
	}, false)
	assert.NoError(t, err)
	assert.Len(t, connections, 1, "project2 should have 1 connection")
	assert.Equal(t, "project2_connection", connections[0].Name)
}

// TestConnectionOrgIDRequired tests that org_id is required for all connections at the SDK level
func TestConnectionOrgIDRequired(t *testing.T) {
	f, p := preConnectionTest(t)

	// Create a valid connection first
	c := f.newConnection(p)
	f.createConnectionsAndAssert(t, c)

	// Parse it to SDK type
	conn, err := scheme.ParseConnection(c)
	require.NoError(t, err)

	// Try to create a connection with invalid org_id via the SDK method
	invalidConn := conn.WithOrgID(sdktypes.InvalidOrgID)
	err = f.gormdb.CreateConnection(f.ctx, invalidConn)
	assert.Error(t, err, "connection without valid org_id should fail at SDK level")
	assert.Contains(t, err.Error(), "org ID is required")
}
