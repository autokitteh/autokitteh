package dbgorm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var u = sdktypes.NewUser("provider", map[string]string{"email": "foo@bar", "name": "Test User"})

// u2 := sdktypes.NewUser("provider2", map[string]string{"email": "email2", "name": "name2"})

func withUser(ctx context.Context, user sdktypes.User) context.Context {
	return authcontext.SetAuthnUser(ctx, user)
}

func preOwnershipTest(t *testing.T) *dbFixture {
	f := newDBFixture()
	findAndAssertCount[scheme.Ownership](t, f, 0, "") // no ownerships
	return f
}

func TestCreateProjectWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	p := f.newProject()
	f.createProjectsAndAssert(t, p)
	assert.NoError(t, f.gormdb.isUserEntity(f.ctx, p.ProjectID))
}

func TestCreateBuildWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	p := f.newProject()
	f.createProjectsAndAssert(t, p)

	// b.ProjectID is nil (optional), thus no user check on create
	b1 := f.newBuild()
	assert.Nil(t, b1.ProjectID)
	f.saveBuildsAndAssert(t, b1)
	assert.NoError(t, f.gormdb.isUserEntity(f.ctx, b1.BuildID))

	// project created by the same user
	b2 := f.newBuild(p)
	f.saveBuildsAndAssert(t, b2)
	assert.NoError(t, f.gormdb.isUserEntity(f.ctx, b2.BuildID))

	// different user - unathorized to create build for the project owned by another user
	f.ctx = withUser(f.ctx, u)
	assert.ErrorIs(t, f.gormdb.saveBuild(f.ctx, &b2), sdkerrors.ErrUnauthorized)
}

func TestCreateDeploymentWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	p := f.newProject()
	e := f.newEnv(p)
	b := f.newBuild()
	f.createProjectsAndAssert(t, p)
	f.createEnvsAndAssert(t, e)
	f.saveBuildsAndAssert(t, b)

	// d.EnvID is nil, but d.buildID is invalid, e.g. zeros - thus unauthorized
	d1 := f.newDeployment()
	assert.Equal(t, d1.BuildID, sdktypes.UUID{}) // zero value
	assert.ErrorIs(t, f.gormdb.createDeployment(f.ctx, &d1), sdkerrors.ErrUnauthorized)

	// with build owned by the same user
	d2 := f.newDeployment(b)
	f.createDeploymentsAndAssert(t, d2)
	assert.NoError(t, f.gormdb.isUserEntity(f.ctx, d2.DeploymentID))

	// with env owned by the same user, but invalid zero buildID
	d3 := f.newDeployment(e)
	assert.ErrorIs(t, f.gormdb.createDeployment(f.ctx, &d3), sdkerrors.ErrUnauthorized)

	// with both build and env owned by the same user
	d4 := f.newDeployment(b, e)
	f.createDeploymentsAndAssert(t, d4)
	assert.NoError(t, f.gormdb.isUserEntity(f.ctx, d4.DeploymentID))

	// different user
	f.ctx = withUser(f.ctx, u)

	// with build owned by the different user
	d5 := f.newDeployment(b)
	assert.ErrorIs(t, f.gormdb.createDeployment(f.ctx, &d5), sdkerrors.ErrUnauthorized)

	// with build and env owned by the different user
	d6 := f.newDeployment(b, e)
	assert.ErrorIs(t, f.gormdb.createDeployment(f.ctx, &d6), sdkerrors.ErrUnauthorized)
}

func TestCreateEnvWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	p := f.newProject()
	f.createProjectsAndAssert(t, p)

	// e.ProjectID is invalid, e.g. zeros - thus unauthorized
	e1 := f.newEnv()
	assert.Equal(t, e1.ProjectID, sdktypes.UUID{}) // zero value
	assert.ErrorIs(t, f.gormdb.createEnv(f.ctx, &e1), sdkerrors.ErrUnauthorized)

	// project created by the same user
	e2 := f.newEnv(p)
	f.createEnvsAndAssert(t, e2)
	assert.NoError(t, f.gormdb.isUserEntity(f.ctx, e2.EnvID))

	// different user - unathorized to create env for the project owned by another user
	f.ctx = withUser(f.ctx, u)
	assert.ErrorIs(t, f.gormdb.createEnv(f.ctx, &e2), sdkerrors.ErrUnauthorized)
}

func TestCreateConnectionWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	p := f.newProject()
	f.createProjectsAndAssert(t, p)

	// c.ProjectID is nil (optional), thus no user check on create
	c1 := f.newConnection()
	assert.Nil(t, c1.ProjectID) // nil
	f.createConnectionsAndAssert(t, c1)
	assert.NoError(t, f.gormdb.isUserEntity(f.ctx, c1.ConnectionID))

	// with project owned by the same user
	c2 := f.newConnection(p)
	f.createConnectionsAndAssert(t, c2)
	assert.NoError(t, f.gormdb.isUserEntity(f.ctx, c2.ConnectionID))

	// different user - unathorized to create connection for the project owned by another user
	f.ctx = withUser(f.ctx, u)
	assert.ErrorIs(t, f.gormdb.createConnection(f.ctx, &c2), sdkerrors.ErrUnauthorized)
}

func TestCreateSessionWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	p := f.newProject()
	b := f.newBuild()
	d := f.newDeployment(b)
	env := f.newEnv(p)
	evt := f.newEvent()

	f.createProjectsAndAssert(t, p)
	f.saveBuildsAndAssert(t, b)
	f.createDeploymentsAndAssert(t, d)
	f.createEnvsAndAssert(t, env)
	f.createEventsAndAssert(t, evt)

	// *BuildID, *EnvID, *DeploymentID, *EventID are null, thus no user check on create
	s1 := f.newSession()
	assert.Nil(t, s1.BuildID, s1.DeploymentID, s1.EnvID, s1.EventID)
	f.createSessionsAndAssert(t, s1)
	assert.NoError(t, f.gormdb.isUserEntity(f.ctx, s1.SessionID))

	// with build owned by the same user
	s2 := f.newSession(b)
	assert.Nil(t, s2.DeploymentID, s2.EnvID, s2.EventID)
	f.createSessionsAndAssert(t, s2)
	assert.NoError(t, f.gormdb.isUserEntity(f.ctx, s2.SessionID))

	// with deployment and build owned by the same user
	s3 := f.newSession(b, d)
	assert.Nil(t, s3.EnvID, s3.EventID)
	f.createSessionsAndAssert(t, s3)
	assert.NoError(t, f.gormdb.isUserEntity(f.ctx, s3.SessionID))

	// with build, deployment and env owned by the same user
	s4 := f.newSession(b, d, env)
	assert.Nil(t, s4.EventID)
	f.createSessionsAndAssert(t, s4)
	assert.NoError(t, f.gormdb.isUserEntity(f.ctx, s4.SessionID))

	// with everything owned by the same user
	s5 := f.newSession(b, d, env, evt)
	f.createSessionsAndAssert(t, s5)
	assert.NoError(t, f.gormdb.isUserEntity(f.ctx, s5.SessionID))

	// different user
	f.ctx = withUser(f.ctx, u)

	// all IDs are nil - could create
	s6 := f.newSession()
	f.createSessionsAndAssert(t, s6)
	assert.NoError(t, f.gormdb.isUserEntity(f.ctx, s6.SessionID))

	// if even one of entities owned by another user - unauthorized
	s7 := f.newSession(b)
	assert.ErrorIs(t, f.gormdb.createSession(f.ctx, &s7), sdkerrors.ErrUnauthorized)

	// and this cannot be overrided by providing one of the entities owned by the same user
	b2 := f.newBuild()
	f.saveBuildsAndAssert(t, b2)
	s8 := f.newSession(b2, d)
	assert.ErrorIs(t, f.gormdb.createSession(f.ctx, &s8), sdkerrors.ErrUnauthorized)
}

func TestCreateEventWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	p := f.newProject()
	c := f.newConnection(p)

	f.createProjectsAndAssert(t, p)
	f.createConnectionsAndAssert(t, c)

	// e.ConnectionID is nil (optional), thus no user check on create
	e1 := f.newEvent()
	assert.Nil(t, e1.ConnectionID)
	f.createEventsAndAssert(t, e1)
	assert.NoError(t, f.gormdb.isUserEntity(f.ctx, e1.EventID))

	// with connection owned by the same user
	e2 := f.newEvent(c)
	f.createEventsAndAssert(t, e2)
	assert.NoError(t, f.gormdb.isUserEntity(f.ctx, e2.EventID))

	// different user - unathorized to create event with connection owned by another user
	f.ctx = withUser(f.ctx, u)
	assert.ErrorIs(t, f.gormdb.saveEvent(f.ctx, &e2), sdkerrors.ErrUnauthorized)
}

func TestCreateTriggerWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	p := f.newProject()
	e := f.newEnv(p)
	c := f.newConnection(p)

	f.createProjectsAndAssert(t, p)
	f.createEnvsAndAssert(t, e)
	f.createConnectionsAndAssert(t, c)

	connCron := f.newConnection()
	connCron.ConnectionID = sdktypes.BuiltinSchedulerConnectionID.UUIDValue()
	f.createConnectionsAndAssert(t, connCron)

	// t.ProjectID, t.EnvID and t.ConnectionID are zeros - thus unauthorized
	t1 := f.newTrigger()
	assert.Equal(t, t1.ProjectID, sdktypes.UUID{})    // zero
	assert.Equal(t, t1.EnvID, sdktypes.UUID{})        // zero
	assert.Equal(t, t1.ConnectionID, sdktypes.UUID{}) // zero
	assert.ErrorIs(t, f.gormdb.createTrigger(f.ctx, &t1), sdkerrors.ErrUnauthorized)

	// user owned projectID and EnvID, but default/zero connection - still unauthorized
	t2 := f.newTrigger(p, e)
	assert.ErrorIs(t, f.gormdb.createTrigger(f.ctx, &t2), sdkerrors.ErrUnauthorized)
	// change connection to builtin scheduler connection (now user owned) - should be allowed
	t2.ConnectionID = connCron.ConnectionID
	f.createTriggersAndAssert(t, t2)
	assert.NoError(t, f.gormdb.isUserEntity(f.ctx, t2.TriggerID))

	t3 := f.newTrigger(p, e, c)
	f.createTriggersAndAssert(t, t3)
	assert.NoError(t, f.gormdb.isUserEntity(f.ctx, t3.TriggerID))

	// different user
	f.ctx = withUser(f.ctx, u)

	// user not owning any of project, env, connecion
	t4 := f.newTrigger(p, e, c)
	assert.ErrorIs(t, f.gormdb.createTrigger(f.ctx, &t4), sdkerrors.ErrUnauthorized)

	// user is not owning only env (common connection and project owned by user) - still unauthorized
	p2 := f.newProject()
	f.createProjectsAndAssert(t, p2)
	t5 := f.newTrigger(p2, e, connCron)
	assert.ErrorIs(t, f.gormdb.createTrigger(f.ctx, &t5), sdkerrors.ErrUnauthorized)
}

func TestCreateVarWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	c, env := createConnectionAndEnv(t, f)

	// env scoped var
	v1 := f.newVar("k", "v", env)
	f.setVarsAndAssert(t, v1)
	assert.NoError(t, f.gormdb.isUserEntity(f.ctx, v1.ScopeID))

	// connection scoped var
	v2 := f.newVar("k", "v", c)
	f.setVarsAndAssert(t, v2)
	assert.NoError(t, f.gormdb.isUserEntity(f.ctx, v2.ScopeID))

	// different user
	f.ctx = withUser(f.ctx, u)

	// cannot create var for non-user owned scope
	v3 := f.newVar("k", "v", env)
	assert.ErrorIs(t, f.gormdb.setVar(f.ctx, &v3), sdkerrors.ErrUnauthorized)

	v4 := f.newVar("k", "v", c)
	assert.ErrorIs(t, f.gormdb.setVar(f.ctx, &v4), sdkerrors.ErrUnauthorized)
}

func TestGetProjectWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	p := f.newProject()
	f.createProjectsAndAssert(t, p)

	// Project owned by the same user tested in TestGetProject

	// different user
	f.ctx = withUser(f.ctx, u)

	_, err := f.gormdb.getProject(f.ctx, p.ProjectID)
	assert.Error(t, err, sdkerrors.ErrUnauthorized)

	_, err = f.gormdb.getProjectByName(f.ctx, p.Name)
	assert.Error(t, err, sdkerrors.ErrUnauthorized)
}

func TestGetBuildWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	b := f.newBuild()

	// Build owned by the same user tested in TestGetBuild

	// different user
	f.ctx = withUser(f.ctx, u)

	_, err := f.gormdb.getBuild(f.ctx, b.BuildID)
	assert.Error(t, err, sdkerrors.ErrUnauthorized)
}

func TestGetDeploymentWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	_, d := createBuildAndDeployment(t, f)

	// Deployment owned by the same user tested in TestGetDeployment

	// different user
	f.ctx = withUser(f.ctx, u)

	_, err := f.gormdb.getDeployment(f.ctx, d.DeploymentID)
	assert.Error(t, err, sdkerrors.ErrUnauthorized)
}

func TestGetEnvWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	p, e := createProjectAndEnv(t, f)

	// Env owned by the same user tested in TestGetEnv

	// different user
	f.ctx = withUser(f.ctx, u)

	_, err := f.gormdb.getEnvByID(f.ctx, e.EnvID)
	assert.Error(t, err, sdkerrors.ErrUnauthorized)

	_, err = f.gormdb.getEnvByName(f.ctx, p.ProjectID, e.Name)
	assert.Error(t, err, sdkerrors.ErrUnauthorized)
}

func TestGetConnectionWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	c := f.newConnection()
	f.createConnectionsAndAssert(t, c)

	// Connection owned by the same user tested in TestGetConnection

	// different user
	f.ctx = withUser(f.ctx, u)

	_, err := f.gormdb.getConnection(f.ctx, c.ConnectionID)
	assert.Error(t, err, sdkerrors.ErrUnauthorized)
}

func TestGetSessionWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	s := f.newSession(sdktypes.SessionStateTypeCompleted)
	f.createSessionsAndAssert(t, s)

	// Session owned by the same user tested in TestGetSession

	// different user
	f.ctx = withUser(f.ctx, u)

	_, err := f.gormdb.getSession(f.ctx, s.SessionID)
	assert.Error(t, err, sdkerrors.ErrUnauthorized)
}

func TestGetEventWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	e := f.newEvent()
	f.createEventsAndAssert(t, e)

	// Event owned by the same user tested in TestGetEvent

	// different user
	f.ctx = withUser(f.ctx, u)

	_, err := f.gormdb.getEvent(f.ctx, e.EventID)
	assert.Error(t, err, sdkerrors.ErrUnauthorized)
}

func TestGetTriggerWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	p, c, e := createProjectConnectionEnv(t, f)
	trg := f.newTrigger(p, c, e)
	f.createTriggersAndAssert(t, trg)

	// Trigger owned by the same user tested in TestGetTrigger

	// different user
	f.ctx = withUser(f.ctx, u)

	_, err := f.gormdb.getEvent(f.ctx, trg.TriggerID)
	assert.Error(t, err, sdkerrors.ErrUnauthorized)
}

func TestListVarsWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)
	c, env := createConnectionAndEnv(t, f)

	v1 := f.newVar("scope", "env", env)
	f.setVarsAndAssert(t, v1)

	v2 := f.newVar("scope", "connection", c)
	f.setVarsAndAssert(t, v2)

	// Var with scope owned by the same user tested in TestListVars

	// different user
	f.ctx = withUser(f.ctx, u)

	vars, err := f.gormdb.listVars(f.ctx, v1.ScopeID, v1.Name)
	assert.Len(t, vars, 0) // no vars fetched, since not not user owned
	assert.NoError(t, err)

	vars, err = f.gormdb.listVars(f.ctx, v2.ScopeID, v2.Name)
	assert.Len(t, vars, 0) // no vars fetched, since not not user owned
	assert.NoError(t, err)
}
