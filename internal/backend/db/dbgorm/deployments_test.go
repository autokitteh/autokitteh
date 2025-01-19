package dbgorm

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (f *dbFixture) createDeploymentsAndAssert(t *testing.T, deployments ...scheme.Deployment) {
	for _, deployment := range deployments {
		assert.NoError(t, f.gormdb.createDeployment(f.ctx, &deployment))
		findAndAssertOne(t, f, deployment, "deployment_id = ?", deployment.DeploymentID)
	}
}

func (f *dbFixture) listDeploymentsAndAssert(t *testing.T, expected int) []scheme.Deployment {
	flt := sdkservices.ListDeploymentsFilter{
		State:               sdktypes.DeploymentStateUnspecified,
		Limit:               0,
		IncludeSessionStats: false,
	}

	deployments, err := f.gormdb.listDeployments(f.ctx, flt)
	assert.NoError(t, err)
	assert.Equal(t, expected, len(deployments))
	return deployments
}

func (f *dbFixture) createProjectBuild(t *testing.T) (scheme.Project, scheme.Build) {
	p := f.newProject()
	b := f.newBuild(p)
	f.createProjectsAndAssert(t, p)
	f.saveBuildsAndAssert(t, b)
	return p, b
}

func listDeploymentsWithStatsAndAssert(t *testing.T, f *dbFixture, expected int) []scheme.DeploymentWithStats {
	flt := sdkservices.ListDeploymentsFilter{
		State:               sdktypes.DeploymentStateUnspecified,
		Limit:               0,
		IncludeSessionStats: true,
	}

	deployments, err := f.gormdb.listDeploymentsWithStats(f.ctx, flt)
	assert.NoError(t, err)
	assert.Equal(t, expected, len(deployments))
	return deployments
}

func (f *dbFixture) assertDeploymentsDeleted(t *testing.T, deployments ...scheme.Deployment) {
	for _, deployment := range deployments {
		assertDeleted(t, f, deployment)
	}
}

func (f *dbFixture) assertDeploymentState(t *testing.T, id uuid.UUID, state int32) {
	d, err := f.gormdb.getDeployment(f.ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, state, d.State)
}

func preDeploymentTest(t *testing.T) *dbFixture {
	f := newDBFixture()
	f.listDeploymentsAndAssert(t, 0) // no deployments
	return f
}

func createBuildAndDeployment(t *testing.T, f *dbFixture, p scheme.Project) (scheme.Build, scheme.Deployment) {
	b := f.newBuild(p)
	f.createProjectsAndAssert(t, p)

	d := f.newDeployment(b, p)
	f.saveBuildsAndAssert(t, b)
	f.createDeploymentsAndAssert(t, d)

	return b, d
}

func TestCreateDeployment(t *testing.T) {
	f := preDeploymentTest(t)

	_, _ = createBuildAndDeployment(t, f, f.newProject())
}

func TestCreateDeploymentsForeignKeys(t *testing.T) {
	// check session creation if foreign keys are not nil
	f := preDeploymentTest(t)

	p, b := f.createProjectBuild(t)

	// valid env, but zero buildID
	d2 := f.newDeployment(p)
	assert.Equal(t, d2.BuildID, uuid.UUID{}) // zero value for buildID
	assert.ErrorIs(t, f.gormdb.createDeployment(f.ctx, &d2), gorm.ErrForeignKeyViolated)

	// use existing user-owned buildID as fake unexisting projectID
	d3 := f.newDeployment(b)
	d3.ProjectID, _ = uuid.NewV7() // no such projectID, since it's a buildID
	assert.ErrorIs(t, f.gormdb.createDeployment(f.ctx, &d3), gorm.ErrForeignKeyViolated)

	// test with existing assets
	d3.ProjectID = p.ProjectID
	f.createDeploymentsAndAssert(t, d3)
}

func TestGetDeployment(t *testing.T) {
	f := preDeploymentTest(t)

	_, d := createBuildAndDeployment(t, f, f.newProject())

	// check getDeployment
	d2, err := f.gormdb.getDeployment(f.ctx, d.DeploymentID)
	if assert.NoError(t, err) {
		resetTimes(d2)
		assert.Equal(t, d, *d2)
	}

	// check getDeployment after delete
	assert.NoError(t, f.gormdb.deleteDeployment(f.ctx, d.DeploymentID))
	_, err = f.gormdb.getDeployment(f.ctx, d.DeploymentID)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestListDeployments(t *testing.T) {
	f := preDeploymentTest(t)
	_, d := createBuildAndDeployment(t, f, f.newProject())

	deployments := f.listDeploymentsAndAssert(t, 1)
	resetTimes(&deployments[0])
	assert.Equal(t, d, deployments[0])

	// test listDeployments after delete
	assert.NoError(t, f.gormdb.deleteDeployment(f.ctx, d.DeploymentID))
	f.listDeploymentsAndAssert(t, 0)
}

func TestListDeploymentsWithStats(t *testing.T) {
	f := preDeploymentTest(t)

	// create deployment and ensure there are no stats
	p := f.newProject()
	b, d := createBuildAndDeployment(t, f, p)

	dWS := scheme.DeploymentWithStats{Deployment: d} // no stats, all zeros
	deployments := listDeploymentsWithStatsAndAssert(t, f, 1)
	resetTimes(&deployments[0], &dWS.Deployment)
	assert.Equal(t, dWS, deployments[0])

	// add session for the stats
	s := f.newSession(sdktypes.SessionStateTypeCompleted, d, b, p)
	f.createSessionsAndAssert(t, s)

	// ensure that new session is included in stats
	dWS.Completed = 1
	deployments = listDeploymentsWithStatsAndAssert(t, f, 1)
	resetTimes(&deployments[0])
	assert.Equal(t, dWS, deployments[0])

	// delete session
	assert.NoError(t, f.gormdb.deleteSession(f.ctx, s.SessionID))
	f.assertSessionsDeleted(t, s)

	// check that deployment stats are updated
	deployments = listDeploymentsWithStatsAndAssert(t, f, 1)
	dWS.Completed = 0 // completed session was deleted
	resetTimes(&deployments[0])
	assert.Equal(t, dWS, deployments[0])
}

func TestDeleteDeployment(t *testing.T) {
	f := preDeploymentTest(t)

	p := f.newProject()

	b, d := createBuildAndDeployment(t, f, p)

	// add sessions and check that deployment stats are updated
	s1 := f.newSession(sdktypes.SessionStateTypeCompleted, d, p, b)
	s2 := f.newSession(sdktypes.SessionStateTypeError, d, p, b)
	f.createSessionsAndAssert(t, s1, s2)

	dWS := scheme.DeploymentWithStats{Deployment: d, Completed: 1, Error: 1}
	deployments := listDeploymentsWithStatsAndAssert(t, f, 1)
	resetTimes(&deployments[0])
	assert.Equal(t, dWS, deployments[0])

	assert.NoError(t, f.gormdb.deleteDeployment(f.ctx, d.DeploymentID))
	f.assertDeploymentsDeleted(t, d)

	listDeploymentsWithStatsAndAssert(t, f, 0)
	f.assertSessionsDeleted(t, s1, s2)

	// TODO: meanwhile builds are not deleted when deployment is deleted
}

func TestUpdateDeploymentStateReturning(t *testing.T) {
	f := preDeploymentTest(t)

	_, d := createBuildAndDeployment(t, f, f.newProject())

	prevState := sdktypes.DeploymentStateUnspecified
	f.assertDeploymentState(t, d.DeploymentID, int32(prevState.ToProto()))

	states := []sdktypes.DeploymentState{
		sdktypes.DeploymentStateActive, sdktypes.DeploymentStateDraining, sdktypes.DeploymentStateInactive,
	}
	for i := range states {
		newState := states[i]
		prevStateFromUpdate, err := f.gormdb.updateDeploymentState(f.ctx, d.DeploymentID, newState)
		assert.NoError(t, err)
		f.assertDeploymentState(t, d.DeploymentID, int32(newState.ToProto()))
		assert.Equal(t, prevState, prevStateFromUpdate)
		prevState = newState
	}
}

func setupDrainingTests(t *testing.T) (*dbFixture, []sdktypes.DeploymentID) {
	f := newDBFixture()

	ps := []scheme.Project{f.newProject(), f.newProject(), f.newProject()}
	ds := make([]scheme.Deployment, 3)
	bs := make([]scheme.Build, 3)

	bs[0], ds[0] = createBuildAndDeployment(t, f, ps[0])
	f.createSessionsAndAssert(t, f.newSession(sdktypes.SessionStateTypeCompleted, ds[0], ps[0], bs[0]))
	f.createSessionsAndAssert(t, f.newSession(sdktypes.SessionStateTypeCompleted, ds[0], ps[0], bs[0]))
	f.createSessionsAndAssert(t, f.newSession(sdktypes.SessionStateTypeCompleted, ds[0], ps[0], bs[0]))

	_, err := f.gormdb.updateDeploymentState(f.ctx, ds[0].DeploymentID, sdktypes.DeploymentStateDraining)
	require.NoError(t, err)

	bs[1], ds[1] = createBuildAndDeployment(t, f, ps[1])
	f.createSessionsAndAssert(t, f.newSession(sdktypes.SessionStateTypeCompleted, ds[1], ps[1], bs[1]))
	f.createSessionsAndAssert(t, f.newSession(sdktypes.SessionStateTypeRunning, ds[1], ps[1], bs[1]))
	f.createSessionsAndAssert(t, f.newSession(sdktypes.SessionStateTypeCompleted, ds[1], ps[1], bs[1]))

	_, err = f.gormdb.updateDeploymentState(f.ctx, ds[1].DeploymentID, sdktypes.DeploymentStateDraining)
	require.NoError(t, err)

	bs[2], ds[2] = createBuildAndDeployment(t, f, ps[2])
	f.createSessionsAndAssert(t, f.newSession(sdktypes.SessionStateTypeCompleted, ds[2], ps[2], bs[2]))
	f.createSessionsAndAssert(t, f.newSession(sdktypes.SessionStateTypeError, ds[2], ps[2], bs[2]))
	f.createSessionsAndAssert(t, f.newSession(sdktypes.SessionStateTypeCompleted, ds[2], ps[2], bs[2]))

	_, err = f.gormdb.updateDeploymentState(f.ctx, ds[2].DeploymentID, sdktypes.DeploymentStateActive)
	require.NoError(t, err)

	deps, err := f.gormdb.listDeployments(f.ctx, sdkservices.ListDeploymentsFilter{})
	require.NoError(t, err)
	require.Equal(t, int32(sdktypes.DeploymentStateDraining.ToProto()), deps[2].State)
	require.Equal(t, int32(sdktypes.DeploymentStateDraining.ToProto()), deps[1].State)
	require.Equal(t, int32(sdktypes.DeploymentStateActive.ToProto()), deps[0].State)

	return f, []sdktypes.DeploymentID{
		sdktypes.NewIDFromUUID[sdktypes.DeploymentID](ds[0].DeploymentID),
		sdktypes.NewIDFromUUID[sdktypes.DeploymentID](ds[1].DeploymentID),
		sdktypes.NewIDFromUUID[sdktypes.DeploymentID](ds[2].DeploymentID),
	}
}

func TestDeactivateAllDrainedDeployments(t *testing.T) {
	f, _ := setupDrainingTests(t)

	n, err := f.gormdb.DeactivateAllDrainedDeployments(f.ctx)
	if assert.NoError(t, err) {
		assert.Equal(t, 1, n)
	}

	deps, err := f.gormdb.listDeployments(f.ctx, sdkservices.ListDeploymentsFilter{})
	if assert.NoError(t, err) {
		assert.Equal(t, int32(sdktypes.DeploymentStateInactive.ToProto()), deps[2].State)
		assert.Equal(t, int32(sdktypes.DeploymentStateDraining.ToProto()), deps[1].State)
		assert.Equal(t, int32(sdktypes.DeploymentStateActive.ToProto()), deps[0].State)
	}
}

func TestDeactivateDrainedDeployments(t *testing.T) {
	f, ds := setupDrainingTests(t)

	wasDrained, err := f.gormdb.DeactivateDrainedDeployment(f.ctx, ds[0])
	if assert.NoError(t, err) {
		assert.True(t, wasDrained)
	}

	wasDrained, err = f.gormdb.DeactivateDrainedDeployment(f.ctx, ds[1])
	if assert.NoError(t, err) {
		assert.False(t, wasDrained)
	}

	wasDrained, err = f.gormdb.DeactivateDrainedDeployment(f.ctx, ds[2])
	if assert.NoError(t, err) {
		assert.False(t, wasDrained)
	}

	deps, err := f.gormdb.listDeployments(f.ctx, sdkservices.ListDeploymentsFilter{})
	if assert.NoError(t, err) {
		assert.Equal(t, int32(sdktypes.DeploymentStateInactive.ToProto()), deps[2].State)
		assert.Equal(t, int32(sdktypes.DeploymentStateDraining.ToProto()), deps[1].State)
		assert.Equal(t, int32(sdktypes.DeploymentStateActive.ToProto()), deps[0].State)
	}
}
