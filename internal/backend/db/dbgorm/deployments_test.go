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
	f := newDBFixture()
	f.listDeploymentsAndAssert(t, 0) // no deployments
	return f
}

func TestCreateDeployment(t *testing.T) {
	f := preDeploymentTest(t)

	d := f.newDeployment()
	// test createDeployment without any assets deployment depends on, since they are soft-foreign keys and could be nil
	f.createDeploymentsAndAssert(t, d)
}

func TestCreateDeploymentsForeignKeys(t *testing.T) {
	// check session creation if foreign keys are not nil
	f := preDeploymentTest(t)

	// negative test with non-existing assets
	d := f.newDeployment()
	unexisting := "unexisting"

	d.BuildID = &unexisting
	assert.ErrorIs(t, f.gormdb.createDeployment(f.ctx, &d), gorm.ErrForeignKeyViolated)
	d.BuildID = nil

	d.EnvID = &unexisting
	assert.ErrorIs(t, f.gormdb.createDeployment(f.ctx, &d), gorm.ErrForeignKeyViolated)
	d.EnvID = nil

	// test with existing assets
	e := f.newEnv()
	b := f.newBuild()
	f.createEnvsAndAssert(t, e)
	f.saveBuildsAndAssert(t, b)

	d = f.newDeployment()
	d.BuildID = &b.BuildID
	d.EnvID = &e.EnvID
	f.createDeploymentsAndAssert(t, d)
}

func TestGetDeployment(t *testing.T) {
	f := preDeploymentTest(t)

	d := f.newDeployment()
	f.createDeploymentsAndAssert(t, d)

	// check getDeployment
	deployment, err := f.gormdb.getDeployment(f.ctx, d.DeploymentID)
	assert.NoError(t, err)
	assert.Equal(t, d, *deployment)

	assert.NoError(t, f.gormdb.deleteDeployment(f.ctx, d.DeploymentID))
	_, err = f.gormdb.getDeployment(f.ctx, d.DeploymentID)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestListDeployments(t *testing.T) {
	f := preDeploymentTest(t)

	d := f.newDeployment()
	f.createDeploymentsAndAssert(t, d)

	deployments := f.listDeploymentsAndAssert(t, 1)
	assert.Equal(t, d, deployments[0])
}

func TestListDeploymentsWithStats(t *testing.T) {
	f := preDeploymentTest(t)

	// create deployment and ensure there are no stats
	d := f.newDeployment()
	f.createDeploymentsAndAssert(t, d)

	dWS := scheme.DeploymentWithStats{Deployment: d} // no stats, all zeros
	deployments := listDeploymentsWithStatsAndAssert(t, f, 1)
	assert.Equal(t, dWS, deployments[0])

	// add session for the stats
	s := f.newSession(sdktypes.SessionStateTypeCompleted)
	s.DeploymentID = &d.DeploymentID
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

	b := f.newBuild()
	d := f.newDeployment()
	d.BuildID = &b.BuildID
	f.saveBuildsAndAssert(t, b)
	f.createDeploymentsAndAssert(t, d)

	// add sessions and check that deployment stats are updated
	s1 := f.newSession(sdktypes.SessionStateTypeCompleted)
	s2 := f.newSession(sdktypes.SessionStateTypeError)
	s1.DeploymentID = &d.DeploymentID
	s2.DeploymentID = &d.DeploymentID
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
