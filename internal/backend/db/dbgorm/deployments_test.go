package dbgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func createDeploymentsAndAssert(t *testing.T, f *dbFixture, deployments ...scheme.Deployment) {
	for _, deployment := range deployments {
		assert.NoError(t, f.gormdb.createDeployment(f.ctx, &deployment))
		findAndAssertOne(t, f, deployment, "deployment_id = ?", deployment.DeploymentID)
	}
}

func listDeploymentsAndAssert(t *testing.T, f *dbFixture, expected int) []scheme.Deployment {
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

func assertDeploymentsDeleted(t *testing.T, f *dbFixture, deployments ...scheme.Deployment) {
	for _, deployment := range deployments {
		assertSoftDeleted(t, f, scheme.Deployment{DeploymentID: deployment.DeploymentID})
	}
}

func TestCreateDeployment(t *testing.T) {
	f := newDBFixture(true)           // no foreign keys
	listDeploymentsAndAssert(t, f, 0) // no deployments

	d := newDeployment(f)
	// test createDeployment
	createDeploymentsAndAssert(t, f, d)
}

func TestCreateDeploymentsForeignKeys(t *testing.T) {
	f := newDBFixture(false) // with foreign keys
	d := newDeployment(f)
	d.BuildID = "unexistingBuildID"
	d.EnvID = "unexistingEnvID"

	err := f.gormdb.createDeployment(f.ctx, &d)
	assert.ErrorContains(t, err, "FOREIGN KEY")

	p := newProject(f)
	e := newEnv(f)
	b := newBuild()
	createProjectsAndAssert(t, f, p)
	createEnvsAndAssert(t, f, e)
	saveBuildsAndAssert(t, f, b)

	d = newDeployment(f)
	createDeploymentsAndAssert(t, f, d)
}

func TestGetDeployment(t *testing.T) {
	f := newDBFixture(true)           // no foreign keys
	listDeploymentsAndAssert(t, f, 0) // no deployments

	d := newDeployment(f)
	createDeploymentsAndAssert(t, f, d)

	// check getDeployment
	deployment, err := f.gormdb.getDeployment(f.ctx, d.DeploymentID)
	assert.NoError(t, err)
	assert.Equal(t, d, *deployment)

	assert.NoError(t, f.gormdb.deleteDeployment(f.ctx, d.DeploymentID))
	_, err = f.gormdb.getDeployment(f.ctx, d.DeploymentID)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestListDeployments(t *testing.T) {
	f := newDBFixture(true)           // no foreign keys
	listDeploymentsAndAssert(t, f, 0) // no deployments

	d := newDeployment(f)
	createDeploymentsAndAssert(t, f, d)

	deployments := listDeploymentsAndAssert(t, f, 1)
	assert.Equal(t, d, deployments[0])
}

func TestListDeploymentsWithStats(t *testing.T) {
	f := newDBFixture(true)           // no foreign keys
	listDeploymentsAndAssert(t, f, 0) // ensure no deployments

	// create deployment and ensure there are no stats
	d := newDeployment(f)
	createDeploymentsAndAssert(t, f, d)

	dWS := scheme.DeploymentWithStats{Deployment: d} // no stats, all zeros
	deployments := listDeploymentsWithStatsAndAssert(t, f, 1)
	assert.Equal(t, dWS, deployments[0])

	// add session for the stats
	s := newSession(f, sdktypes.SessionStateTypeCompleted)
	createSessionsAndAssert(t, f, s)

	// ensure that new session is included in stats
	dWS.Completed = 1
	deployments = listDeploymentsWithStatsAndAssert(t, f, 1)
	assert.Equal(t, dWS, deployments[0])

	// delete session
	assert.NoError(t, f.gormdb.deleteSession(f.ctx, s.SessionID))
	assertSessionsDeleted(t, f, s)

	// check that deployment stats are updated
	deployments = listDeploymentsWithStatsAndAssert(t, f, 1)
	dWS.Completed = 0 // completed session was deleted
	assert.Equal(t, dWS, deployments[0])
}

func TestDeleteDeployment(t *testing.T) {
	f := newDBFixture(true)           // no foreign keys
	listDeploymentsAndAssert(t, f, 0) // ensure no deployments

	b := newBuild()
	d := newDeployment(f)
	d.BuildID = b.BuildID
	saveBuildsAndAssert(t, f, b)
	createDeploymentsAndAssert(t, f, d)

	// add sessions and check that deployment stats are updated
	s1 := newSession(f, sdktypes.SessionStateTypeCompleted)
	s2 := newSession(f, sdktypes.SessionStateTypeError)
	createSessionsAndAssert(t, f, s1, s2)

	dWS := scheme.DeploymentWithStats{Deployment: d, Completed: 1, Error: 1}
	deployments := listDeploymentsWithStatsAndAssert(t, f, 1)
	assert.Equal(t, dWS, deployments[0])

	// delete deployment. Ensure deployment sessions are marked as deleted as well
	assert.NoError(t, f.gormdb.deleteDeployment(f.ctx, d.DeploymentID))
	assertDeploymentsDeleted(t, f, d)

	listDeploymentsWithStatsAndAssert(t, f, 0)
	assertSessionsDeleted(t, f, s1, s2)

	// TODO: meanwhile builds are not deleted when deployment is deleted
	// assertBuildDeleted(t, f, b.BuildID)
}
