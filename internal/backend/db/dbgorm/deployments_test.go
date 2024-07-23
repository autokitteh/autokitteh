package dbgorm

import (
	"testing"

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

func preDeploymentTest(t *testing.T) *dbFixture {
	f := newDBFixture().withUser(sdktypes.DefaultUser)
	f.listDeploymentsAndAssert(t, 0) // no deployments
	return f
}

func createBuildAndDeployment(t *testing.T, f *dbFixture) (scheme.Build, scheme.Deployment) {
	b := f.newBuild()
	d := f.newDeployment(b)
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

	_, b, e := f.createProjectBuildEnv(t)

	// negative test with non-existing assets
	// zero buildID
	d1 := f.newDeployment()
	assert.Equal(t, d1.BuildID, sdktypes.UUID{}) // zero value for buildID
	assert.ErrorIs(t, f.gormdb.createDeployment(f.ctx, &d1), gorm.ErrForeignKeyViolated)

	// valid env, but zero buildID
	d2 := f.newDeployment(e)
	assert.Equal(t, d2.BuildID, sdktypes.UUID{}) // zero value for buildID
	assert.ErrorIs(t, f.gormdb.createDeployment(f.ctx, &d2), gorm.ErrForeignKeyViolated)

	// use existing user-owned buildID as fake unexisting envID
	d3 := f.newDeployment(b)
	d3.EnvID = &b.BuildID // no such envID, since it's a buildID
	assert.ErrorIs(t, f.gormdb.createDeployment(f.ctx, &d3), gorm.ErrForeignKeyViolated)

	// test with existing assets
	d3.EnvID = &e.EnvID
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
