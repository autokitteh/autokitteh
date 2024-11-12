package dbgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	akCtx "go.autokitteh.dev/autokitteh/internal/backend/context"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	u1 = sdktypes.NewUser("foo@baz", "foobaz").WithNewID()
	u2 = sdktypes.NewUser("foo@bar", "foobar").WithNewID()
)

func withSystemOrginators(t *testing.T, f *dbFixture, test func(t *testing.T, f *dbFixture)) {
	for _, orginator := range akCtx.SystemOrginators {
		t.Run(orginator.String(), func(t *testing.T) {
			f.ctx = akCtx.WithRequestOrginator(f.ctx, orginator)
			test(t, f)
		})
	}
}

func preOwnershipTest(t *testing.T) *dbFixture {
	f := newDBFixture().withUser(u1)
	f.ctx = akCtx.WithRequestOrginator(f.ctx, akCtx.User)
	findAndAssertCount[scheme.Ownership](t, f, 0, "") // no ownerships
	return f
}

func (f *dbFixture) createProjectBuild(t *testing.T) (scheme.Project, scheme.Build) {
	p := f.newProject()
	b := f.newBuild()
	f.createProjectsAndAssert(t, p)
	f.saveBuildsAndAssert(t, b)
	return p, b
}

func TestCreateWithoutUser(t *testing.T) {
	// creation of any object with ownership requires user
	f := newDBFixture() // no user

	p := f.newProject()
	assert.ErrorContains(t, f.gormdb.createProject(f.ctx, &p), "unknown user")

	b := f.newBuild()
	assert.ErrorContains(t, f.gormdb.saveBuild(f.ctx, &b), "unknown user")

	d := f.newDeployment()
	assert.ErrorContains(t, f.gormdb.createDeployment(f.ctx, &d), "unknown user")

	c := f.newConnection()
	assert.ErrorContains(t, f.gormdb.createConnection(f.ctx, &c), "unknown user")

	s := f.newSession()
	assert.ErrorContains(t, f.gormdb.createSession(f.ctx, &s), "unknown user")

	evt := f.newEvent()
	assert.ErrorContains(t, f.gormdb.saveEvent(f.ctx, &evt), "unknown user")

	trg := f.newTrigger()
	assert.ErrorContains(t, f.gormdb.createTrigger(f.ctx, &trg), "unknown user")
}

func TestCreateProjectWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	p := f.newProject()
	f.createProjectsAndAssert(t, p)
	assert.NoError(t, f.gormdb.isCtxUserEntity(f.ctx, p.ProjectID))
}

func TestCreateBuildWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	p := f.newProject()
	f.createProjectsAndAssert(t, p)

	// b.ProjectID is nil (optional), thus no user check on create
	b1 := f.newBuild()
	assert.Nil(t, b1.ProjectID)
	f.saveBuildsAndAssert(t, b1)
	assert.NoError(t, f.gormdb.isCtxUserEntity(f.ctx, b1.BuildID))

	// project created by the same user
	b2 := f.newBuild(p)
	f.saveBuildsAndAssert(t, b2)
	assert.NoError(t, f.gormdb.isCtxUserEntity(f.ctx, b2.BuildID))

	// different user - unathorized to create build for the project owned by another user
	f.withUser(u2)
	assert.ErrorIs(t, f.gormdb.saveBuild(f.ctx, &b2), sdkerrors.ErrUnauthorized)
}

func TestCreateDeploymentWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	p, b := f.createProjectBuild(t)

	// with build owned by the same user
	d1 := f.newDeployment(b)
	f.createDeploymentsAndAssert(t, d1)
	assert.NoError(t, f.gormdb.isCtxUserEntity(f.ctx, d1.DeploymentID))

	// with both build and env owned by the same user
	d2 := f.newDeployment(b, p)
	f.createDeploymentsAndAssert(t, d2)
	assert.NoError(t, f.gormdb.isCtxUserEntity(f.ctx, d2.DeploymentID))

	// different user
	f.withUser(u2)

	// with build owned by the different user
	d4 := f.newDeployment(b)
	assert.ErrorIs(t, f.gormdb.createDeployment(f.ctx, &d4), sdkerrors.ErrUnauthorized)

	// with build and env owned by the different user
	d5 := f.newDeployment(b, p)
	assert.ErrorIs(t, f.gormdb.createDeployment(f.ctx, &d5), sdkerrors.ErrUnauthorized)
}

func TestCreateConnectionWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	p := f.newProject()
	f.createProjectsAndAssert(t, p)

	// c.ProjectID is nil (optional), thus no user check on create
	c1 := f.newConnection()
	assert.Nil(t, c1.ProjectID) // nil
	f.createConnectionsAndAssert(t, c1)
	assert.NoError(t, f.gormdb.isCtxUserEntity(f.ctx, c1.ConnectionID))

	// with project owned by the same user
	c2 := f.newConnection(p)
	f.createConnectionsAndAssert(t, c2)
	assert.NoError(t, f.gormdb.isCtxUserEntity(f.ctx, c2.ConnectionID))

	// different user - unathorized to create connection for the project owned by another user
	f.withUser(u2)
	assert.ErrorIs(t, f.gormdb.createConnection(f.ctx, &c2), sdkerrors.ErrUnauthorized)
}

func TestCreateSessionWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	p := f.newProject()
	b := f.newBuild()
	d := f.newDeployment(b)
	evt := f.newEvent()

	f.createProjectsAndAssert(t, p)
	f.saveBuildsAndAssert(t, b)
	f.createDeploymentsAndAssert(t, d)
	f.createEventsAndAssert(t, evt)

	// *BuildID, *EnvID, *DeploymentID, *EventID are null, thus no user check on create
	s1 := f.newSession()
	assert.Nil(t, s1.BuildID, s1.DeploymentID, s1.EventID)
	f.createSessionsAndAssert(t, s1)
	assert.NoError(t, f.gormdb.isCtxUserEntity(f.ctx, s1.SessionID))

	// with build owned by the same user
	s2 := f.newSession(b)
	assert.Nil(t, s2.DeploymentID, s2.EventID)
	f.createSessionsAndAssert(t, s2)
	assert.NoError(t, f.gormdb.isCtxUserEntity(f.ctx, s2.SessionID))

	// with deployment and build owned by the same user
	s3 := f.newSession(b, d)
	assert.Nil(t, s3.EventID)
	f.createSessionsAndAssert(t, s3)
	assert.NoError(t, f.gormdb.isCtxUserEntity(f.ctx, s3.SessionID))

	// with build, deployment and env owned by the same user
	s4 := f.newSession(b, d)
	assert.Nil(t, s4.EventID)
	f.createSessionsAndAssert(t, s4)
	assert.NoError(t, f.gormdb.isCtxUserEntity(f.ctx, s4.SessionID))

	// with everything owned by the same user
	s5 := f.newSession(b, d, evt)
	f.createSessionsAndAssert(t, s5)
	assert.NoError(t, f.gormdb.isCtxUserEntity(f.ctx, s5.SessionID))

	// different user
	f.withUser(u2)

	// all IDs are nil - could create
	s6 := f.newSession()
	f.createSessionsAndAssert(t, s6)
	assert.NoError(t, f.gormdb.isCtxUserEntity(f.ctx, s6.SessionID))

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
	assert.NoError(t, f.gormdb.isCtxUserEntity(f.ctx, e1.EventID))

	// with connection owned by the same user
	e2 := f.newEvent(c)
	f.createEventsAndAssert(t, e2)
	assert.NoError(t, f.gormdb.isCtxUserEntity(f.ctx, e2.EventID))

	// different user - unathorized to create event with connection owned by another user
	f.withUser(u2)
	assert.ErrorIs(t, f.gormdb.saveEvent(f.ctx, &e2), sdkerrors.ErrUnauthorized)
}

func TestCreateTriggerWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	p, c := f.createProjectConnection(t)

	t1 := f.newTrigger(p, c)
	f.createTriggersAndAssert(t, t1)
	assert.NoError(t, f.gormdb.isCtxUserEntity(f.ctx, t1.TriggerID))

	// different user
	f.withUser(u2)

	// user not owning any of project, env, connecion
	t2 := f.newTrigger(p, c)
	assert.ErrorIs(t, f.gormdb.createTrigger(f.ctx, &t2), sdkerrors.ErrUnauthorized)

	// user is not owning only env (project is owned by user and connection is shared) - still unauthorized
	p2 := f.newProject()
	f.createProjectsAndAssert(t, p2)
}

func TestSetVarWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	c, p := createConnection(t, f)

	// env scoped var
	v1 := f.newVar("k", "v", p)
	f.setVarsAndAssert(t, v1)
	assert.NoError(t, f.gormdb.isCtxUserEntity(f.ctx, v1.ScopeID))

	// connection scoped var
	v2 := f.newVar("k", "v", c)
	f.setVarsAndAssert(t, v2)
	assert.NoError(t, f.gormdb.isCtxUserEntity(f.ctx, v2.ScopeID))

	// different user
	f.withUser(u2)

	// cannot create var for non-user owned scope
	v3 := f.newVar("k", "v", p)
	assert.ErrorIs(t, f.gormdb.setVar(f.ctx, &v3), sdkerrors.ErrUnauthorized)

	v4 := f.newVar("k", "v", c)
	assert.ErrorIs(t, f.gormdb.setVar(f.ctx, &v4), sdkerrors.ErrUnauthorized)
}

func TestDeleteProjectsWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	p := f.newProject()
	f.createProjectsAndAssert(t, p)

	// Project owned by the same user tested in TestDeleteProject

	// different user
	f.withUser(u2)

	assert.Error(t, f.gormdb.deleteProject(f.ctx, p.ProjectID), sdkerrors.ErrUnauthorized)
}

func TestDeleteBuildsWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	b := f.newBuild()
	f.saveBuildsAndAssert(t, b)

	// Project owned by the same user tested in TestDeleteBuild

	// different user
	f.withUser(u2)

	assert.Error(t, f.gormdb.deleteBuild(f.ctx, b.BuildID), sdkerrors.ErrUnauthorized)
}

func TestDeleteDeploymentsWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)
	_, d := createBuildAndDeployment(t, f)

	// Deployment owned by the same user tested in TestDeleteDeployment

	// different user
	f.withUser(u2)

	assert.Error(t, f.gormdb.deleteDeployment(f.ctx, d.DeploymentID), sdkerrors.ErrUnauthorized)
}

func TestDeleteConnectionsWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)
	c := f.newConnection()
	f.createConnectionsAndAssert(t, c)

	// Connection owned by the same user tested in TestDeleteConnection

	// different user
	f.withUser(u2)

	assert.Error(t, f.gormdb.deleteConnection(f.ctx, c.ConnectionID), sdkerrors.ErrUnauthorized)
}

func TestDeleteSessionsWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	s := f.newSession(sdktypes.SessionStateTypeCompleted)
	f.createSessionsAndAssert(t, s)

	// Session owned by the same user tested in TestDeleteSession

	// different user
	f.withUser(u2)

	assert.Error(t, f.gormdb.deleteSession(f.ctx, s.SessionID), sdkerrors.ErrUnauthorized)
}

func TestDeleteEventsWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	e := f.newEvent()
	f.createEventsAndAssert(t, e)

	// Event owned by the same user tested in TestDeleteEvent

	// different user
	f.withUser(u2)

	assert.Error(t, f.gormdb.deleteEvent(f.ctx, e.EventID), sdkerrors.ErrUnauthorized)
}

func TestDeleteTriggersWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	p, c := f.createProjectConnection(t)
	trg := f.newTrigger(p, c)
	f.createTriggersAndAssert(t, trg)

	// Trigger owned by the same user tested in TestDeleteTrigger

	// different user
	f.withUser(u2)

	assert.Error(t, f.gormdb.deleteTrigger(f.ctx, trg.TriggerID), sdkerrors.ErrUnauthorized)
}

func TestDeleteVarsWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)
	c, p := createConnection(t, f)

	v1 := f.newVar("scope", "p", p)
	f.setVarsAndAssert(t, v1)

	v2 := f.newVar("scope", "connection", c)
	f.setVarsAndAssert(t, v2)

	// Var with scope owned by the same user tested in TestDeleteVars

	// different user
	f.withUser(u2)

	assert.Error(t, f.gormdb.deleteVars(f.ctx, v1.ScopeID, v1.Name), sdkerrors.ErrUnauthorized)
	assert.Error(t, f.gormdb.deleteVars(f.ctx, v2.ScopeID, v2.Name), sdkerrors.ErrUnauthorized)
}

func TestGetProjectWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	p := f.newProject()
	f.createProjectsAndAssert(t, p)

	// Project owned by the same user tested in TestGetProject

	// different user
	f.withUser(u2)

	_, err := f.gormdb.getProject(f.ctx, p.ProjectID)
	assert.Error(t, err, sdkerrors.ErrUnauthorized)

	_, err = f.gormdb.getProjectByName(f.ctx, p.Name)
	assert.Error(t, err, sdkerrors.ErrUnauthorized)

	// with system orginators - full access
	withSystemOrginators(t, f, func(t *testing.T, f *dbFixture) {
		_, err := f.gormdb.getProject(f.ctx, p.ProjectID)
		assert.NoError(t, err)

		_, err = f.gormdb.getProjectByName(f.ctx, p.Name)
		assert.NoError(t, err)
	})
}

func TestGetBuildWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	b := f.newBuild()
	f.saveBuildsAndAssert(t, b)

	// Build owned by the same user tested in TestGetBuild

	// different user
	f.withUser(u2)

	_, err := f.gormdb.getBuild(f.ctx, b.BuildID)
	assert.Error(t, err, sdkerrors.ErrUnauthorized)

	// with system orginators - full access
	withSystemOrginators(t, f, func(t *testing.T, f *dbFixture) {
		_, err = f.gormdb.getBuild(f.ctx, b.BuildID)
		assert.NoError(t, err)
	})
}

func TestGetDeploymentWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	_, d := createBuildAndDeployment(t, f)

	// Deployment owned by the same user tested in TestGetDeployment

	// different user
	f.withUser(u2)

	_, err := f.gormdb.getDeployment(f.ctx, d.DeploymentID)
	assert.Error(t, err, sdkerrors.ErrUnauthorized)

	// with system orginators - full access
	withSystemOrginators(t, f, func(t *testing.T, f *dbFixture) {
		_, err = f.gormdb.getDeployment(f.ctx, d.DeploymentID)
		assert.NoError(t, err)
	})
}

func TestGetConnectionWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	c := f.newConnection()
	f.createConnectionsAndAssert(t, c)

	// Connection owned by the same user tested in TestGetConnection

	// different user
	f.withUser(u2)

	_, err := f.gormdb.getConnection(f.ctx, c.ConnectionID)
	assert.Error(t, err, sdkerrors.ErrUnauthorized)

	// with system orginators - full access
	withSystemOrginators(t, f, func(t *testing.T, f *dbFixture) {
		_, err = f.gormdb.getConnection(f.ctx, c.ConnectionID)
		assert.NoError(t, err)
	})
}

func TestGetSessionWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	s := f.newSession(sdktypes.SessionStateTypeCompleted)
	f.createSessionsAndAssert(t, s)

	// Session owned by the same user tested in TestGetSession

	// different user
	f.withUser(u2)

	_, err := f.gormdb.getSession(f.ctx, s.SessionID)
	assert.Error(t, err, sdkerrors.ErrUnauthorized)

	// with system orginators - full access
	withSystemOrginators(t, f, func(t *testing.T, f *dbFixture) {
		_, err = f.gormdb.getSession(f.ctx, s.SessionID)
		assert.NoError(t, err)
	})
}

func TestGetEventWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	e := f.newEvent()
	f.createEventsAndAssert(t, e)

	// Event owned by the same user tested in TestGetEvent

	// different user
	f.withUser(u2)

	_, err := f.gormdb.getEvent(f.ctx, e.EventID)
	assert.Error(t, err, sdkerrors.ErrUnauthorized)

	// with system orginators - full access
	withSystemOrginators(t, f, func(t *testing.T, f *dbFixture) {
		_, err = f.gormdb.getEvent(f.ctx, e.EventID)
		assert.NoError(t, err)
	})
}

func TestGetTriggerWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	p, c := f.createProjectConnection(t)
	trg := f.newTrigger(p, c)
	f.createTriggersAndAssert(t, trg)

	// Trigger owned by the same user tested in TestGetTrigger

	// different user
	f.withUser(u2)

	_, err := f.gormdb.getTriggerByID(f.ctx, trg.TriggerID)
	assert.Error(t, err, sdkerrors.ErrUnauthorized)

	// with system orginators - full access
	withSystemOrginators(t, f, func(t *testing.T, f *dbFixture) {
		_, err = f.gormdb.getTriggerByID(f.ctx, trg.TriggerID)
		assert.NoError(t, err)
	})
}

func TestListProjectsWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	p := f.newProject()
	f.createProjectsAndAssert(t, p)

	// Project owned by the same user tested in TestListProjects

	// different user
	f.withUser(u2)

	projects, err := f.gormdb.listProjects(f.ctx)
	assert.Len(t, projects, 0) // no projects fetched, not user owned
	assert.NoError(t, err)

	// with system orginators - full access
	withSystemOrginators(t, f, func(t *testing.T, f *dbFixture) {
		projects, err = f.gormdb.listProjects(f.ctx)
		assert.Len(t, projects, 1)
		assert.NoError(t, err)
	})
}

func TestListBuildsWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	b := f.newBuild()
	f.saveBuildsAndAssert(t, b)

	// Build owned by the same user tested in TestListBuilds

	// different user
	f.withUser(u2)

	builds, err := f.gormdb.listBuilds(f.ctx, sdkservices.ListBuildsFilter{})
	assert.Len(t, builds, 0) // no build fetched, not user owned
	assert.NoError(t, err)

	// with system orginators - full access
	withSystemOrginators(t, f, func(t *testing.T, f *dbFixture) {
		builds, err = f.gormdb.listBuilds(f.ctx, sdkservices.ListBuildsFilter{})
		assert.Len(t, builds, 1)
		assert.NoError(t, err)
	})
}

func TestListDeploymentsWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	_, _ = createBuildAndDeployment(t, f)

	// Deployment owned by the same user tested in TestListDeployments

	// different user
	f.withUser(u2)

	f.listDeploymentsAndAssert(t, 0) // no deployments fetched, not user owned

	// with system orginators - full access
	withSystemOrginators(t, f, func(t *testing.T, f *dbFixture) {
		f.ctx = akCtx.WithRequestOrginator(f.ctx, akCtx.Dispatcher)
		f.listDeploymentsAndAssert(t, 1)
	})
}

func TestListConnectionsWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	c := f.newConnection()
	f.createConnectionsAndAssert(t, c)

	// Connection owned by the same user tested in TestListConnections

	// different user
	f.withUser(u2)

	cc, err := f.gormdb.listConnections(f.ctx, sdkservices.ListConnectionsFilter{}, false)
	assert.Len(t, cc, 0) // no connections fetched, not user owned
	assert.NoError(t, err)

	// with system orginators - full access
	withSystemOrginators(t, f, func(t *testing.T, f *dbFixture) {
		cc, err = f.gormdb.listConnections(f.ctx, sdkservices.ListConnectionsFilter{}, false)
		assert.Len(t, cc, 1)
		assert.NoError(t, err)
	})
}

func TestListSessionsWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	s := f.newSession(sdktypes.SessionStateTypeCompleted)
	f.createSessionsAndAssert(t, s)

	// Session owned by the same user tested in TestListSessions

	// different user
	f.withUser(u2)

	f.listSessionsAndAssert(t, 0) // no sessions fetched, not user owned

	// with system orginators - full access
	withSystemOrginators(t, f, func(t *testing.T, f *dbFixture) {
		f.listSessionsAndAssert(t, 1)
	})
}

func TestListEventsWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	e := f.newEvent()
	f.createEventsAndAssert(t, e)

	// Event owned by the same user tested in TestListEventsOrder

	// different user
	f.withUser(u2)

	events, err := f.gormdb.listEvents(f.ctx, sdkservices.ListEventsFilter{})
	assert.Len(t, events, 0) // no events fetched, not user owned
	assert.NoError(t, err)

	// with system orginators - full access
	withSystemOrginators(t, f, func(t *testing.T, f *dbFixture) {
		events, err = f.gormdb.listEvents(f.ctx, sdkservices.ListEventsFilter{})
		assert.Len(t, events, 1)
		assert.NoError(t, err)
	})
}

func TestListTriggersWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)

	p, c := f.createProjectConnection(t)
	trg := f.newTrigger(p, c)
	f.createTriggersAndAssert(t, trg)

	// Trigger owned by the same user tested in TestListTriggers

	// different user
	f.withUser(u2)

	triggers, err := f.gormdb.listTriggers(f.ctx, sdkservices.ListTriggersFilter{})
	assert.Len(t, triggers, 0) // no triggers fetched, not user owned
	assert.NoError(t, err)

	// with system orginators - full access
	withSystemOrginators(t, f, func(t *testing.T, f *dbFixture) {
		triggers, err = f.gormdb.listTriggers(f.ctx, sdkservices.ListTriggersFilter{})
		assert.Len(t, triggers, 1)
		assert.NoError(t, err)
	})
}

func TestListVarsWithOwnership(t *testing.T) {
	f := preOwnershipTest(t)
	c, p := createConnection(t, f)

	v1 := f.newVar("scope", "p", p)
	f.setVarsAndAssert(t, v1)

	v2 := f.newVar("scope", "connection", c)
	f.setVarsAndAssert(t, v2)

	// Var with scope owned by the same user tested in TestListVars

	// different user
	f.withUser(u2)

	vars, err := f.gormdb.listVars(f.ctx, v1.ScopeID, v1.Name)
	assert.Len(t, vars, 0) // no vars fetched, since not not user owned
	assert.NoError(t, err)

	vars, err = f.gormdb.listVars(f.ctx, v2.ScopeID, v2.Name)
	assert.Len(t, vars, 0) // no vars fetched, since not not user owned
	assert.NoError(t, err)

	// with system orginators - full access
	withSystemOrginators(t, f, func(t *testing.T, f *dbFixture) {
		vars, err = f.gormdb.listVars(f.ctx, v1.ScopeID, v1.Name)
		assert.Len(t, vars, 1)
		assert.NoError(t, err)

		vars, err = f.gormdb.listVars(f.ctx, v2.ScopeID, v2.Name)
		assert.Len(t, vars, 1)
		assert.NoError(t, err)
	})
}
