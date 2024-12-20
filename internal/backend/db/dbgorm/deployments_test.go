package dbgorm

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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
	b := f.newBuild()
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
		assertSoftDeleted(t, f, deployment)
	}
}

func (f *dbFixture) assertDeploymentState(t *testing.T, id sdktypes.UUID, state int32) {
	d, err := f.gormdb.getDeployment(f.ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, state, d.State)
}

func preDeploymentTest(t *testing.T) *dbFixture {
	f := newDBFixture()
	f.listDeploymentsAndAssert(t, 0) // no deployments
	return f
}

func createBuildAndDeployment(t *testing.T, f *dbFixture) (scheme.Build, scheme.Deployment) {
	b := f.newBuild()
	p := f.newProject()
	f.createProjectsAndAssert(t, p)

	d := f.newDeployment(b, p)
	f.saveBuildsAndAssert(t, b)
	f.createDeploymentsAndAssert(t, d)
	return b, d
}

func TestCreateDeployment(t *testing.T) {
	f := preDeploymentTest(t)

	_, _ = createBuildAndDeployment(t, f)
}

func TestCreateDeploymentsForeignKeys(t *testing.T) {
	// check session creation if foreign keys are not nil
	f := preDeploymentTest(t)

	p, b := f.createProjectBuild(t)

	// negative test with non-existing assets
	// zero buildID
	d1 := f.newDeployment()
	assert.Equal(t, d1.BuildID, sdktypes.UUID{}) // zero value for buildID
	assert.ErrorIs(t, f.gormdb.createDeployment(f.ctx, &d1), gorm.ErrForeignKeyViolated)

	// valid env, but zero buildID
	d2 := f.newDeployment(p)
	assert.Equal(t, d2.BuildID, sdktypes.UUID{}) // zero value for buildID
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

	_, d := createBuildAndDeployment(t, f)

	// check getDeployment
	d2, err := f.gormdb.getDeployment(f.ctx, d.DeploymentID)
	assert.NoError(t, err)
	assert.Equal(t, d, *d2)

	// check getDeployment after delete
	assert.NoError(t, f.gormdb.deleteDeployment(f.ctx, d.DeploymentID))
	_, err = f.gormdb.getDeployment(f.ctx, d.DeploymentID)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestListDeployments(t *testing.T) {
	f := preDeploymentTest(t)
	_, d := createBuildAndDeployment(t, f)

	deployments := f.listDeploymentsAndAssert(t, 1)
	assert.Equal(t, d, deployments[0])

	// test listDeployments after delete
	assert.NoError(t, f.gormdb.deleteDeployment(f.ctx, d.DeploymentID))
	f.listDeploymentsAndAssert(t, 0)
}

func TestListDeploymentsWithStats(t *testing.T) {
	f := preDeploymentTest(t)

	// create deployment and ensure there are no stats
	_, d := createBuildAndDeployment(t, f)

	dWS := scheme.DeploymentWithStats{Deployment: d} // no stats, all zeros
	deployments := listDeploymentsWithStatsAndAssert(t, f, 1)
	assert.Equal(t, dWS, deployments[0])

	// add session for the stats
	s := f.newSession(sdktypes.SessionStateTypeCompleted, d)
	f.createSessionsAndAssert(t, s)

	// ensure that new session is included in stats
	dWS.Completed = 1
	deployments = listDeploymentsWithStatsAndAssert(t, f, 1)
	assert.Equal(t, dWS, deployments[0])

	// delete session
	assert.NoError(t, f.gormdb.deleteSession(f.ctx, s.SessionID))
	f.assertSessionsDeleted(t, s)

	// check that deployment stats are updated
	deployments = listDeploymentsWithStatsAndAssert(t, f, 1)
	dWS.Completed = 0 // completed session was deleted
	assert.Equal(t, dWS, deployments[0])
}

func TestDeleteDeployment(t *testing.T) {
	f := preDeploymentTest(t)

	_, d := createBuildAndDeployment(t, f)

	// add sessions and check that deployment stats are updated
	s1 := f.newSession(sdktypes.SessionStateTypeCompleted, d)
	s2 := f.newSession(sdktypes.SessionStateTypeError, d)
	f.createSessionsAndAssert(t, s1, s2)

	dWS := scheme.DeploymentWithStats{Deployment: d, Completed: 1, Error: 1}
	deployments := listDeploymentsWithStatsAndAssert(t, f, 1)
	assert.Equal(t, dWS, deployments[0])

	// delete deployment. Ensure deployment sessions are marked as deleted as well
	assert.NoError(t, f.gormdb.deleteDeployment(f.ctx, d.DeploymentID))
	f.assertDeploymentsDeleted(t, d)

	listDeploymentsWithStatsAndAssert(t, f, 0)
	f.assertSessionsDeleted(t, s1, s2)

	// TODO: meanwhile builds are not deleted when deployment is deleted
}

/*
func TestDeleteDeploymentForeignKeys(t *testing.T) {
	// deployment is soft-deleted, so no need to check foreign keys meanwhile
}
*/

func TestUpdateDeploymentStateReturning(t *testing.T) {
	f := preDeploymentTest(t)

	_, d := createBuildAndDeployment(t, f)

	prevState := sdktypes.DeploymentStateUnspecified
	f.assertDeploymentState(t, d.DeploymentID, int32(prevState.ToProto()))

	states := []sdktypes.DeploymentState{
		sdktypes.DeploymentStateActive, sdktypes.DeploymentStateDraining, sdktypes.DeploymentStateInactive,
	}
	for i := 0; i < len(states); i++ {
		newState := states[i]
		prevStateFromUpdate, err := f.gormdb.updateDeploymentState(f.ctx, d.DeploymentID, newState)
		assert.NoError(t, err)
		f.assertDeploymentState(t, d.DeploymentID, int32(newState.ToProto()))
		assert.Equal(t, prevState, prevStateFromUpdate)
		prevState = newState
	}
}
