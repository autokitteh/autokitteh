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

func createDeploymentAndTest(t *testing.T, f *dbFixture, deployment scheme.Deployment) {
	assert.NoError(t, f.gormdb.createDeployment(f.ctx, deployment))
	res := f.gormdb.db.First(&scheme.Deployment{}, "deployment_id = ?", deployment.DeploymentID)
	require.NoError(t, res.Error)
	assert.Equal(t, int64(1), res.RowsAffected)
}

func listDeploymentsAndTest(t *testing.T, f *dbFixture, expected int) []scheme.Deployment {
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

func listDeploymentsWithStatsAndTest(t *testing.T, f *dbFixture, expected int) []scheme.DeploymentWithStats {
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

func TestCreateDeployment(t *testing.T) {
	f := newDbFixture(true)         // no foreign keys
	listDeploymentsAndTest(t, f, 0) // ensure no deployments

	d := newDeployment(testBuildID, testEnvID)
	createDeploymentAndTest(t, f, d)

	// obtain all deployments records from the deployments table
	var deployments []scheme.Deployment
	assert.NoError(t, f.db.Find(&deployments).Error)
	assert.Equal(t, int(1), len(deployments))
	assert.Equal(t, d, deployments[0])
}

func TestGetDeployment(t *testing.T) {
	f := newDbFixture(true)         // no foreign keys
	listDeploymentsAndTest(t, f, 0) // ensure no deployments

	d := newDeployment(testBuildID, testEnvID)
	createDeploymentAndTest(t, f, d)

	deployment, err := f.gormdb.getDeployment(f.ctx, d.DeploymentID)
	assert.NoError(t, err)
	assert.Equal(t, d, *deployment)

	assert.NoError(t, f.gormdb.deleteDeployment(f.ctx, d.DeploymentID))
	_, err = f.gormdb.getDeployment(f.ctx, d.DeploymentID)
	assert.Equal(t, err, gorm.ErrRecordNotFound)
}

func TestListDeployments(t *testing.T) {
	f := newDbFixture(true)         // no foreign keys
	listDeploymentsAndTest(t, f, 0) // ensure no deployments

	d := newDeployment(testBuildID, testEnvID)
	createDeploymentAndTest(t, f, d)

	deployments := listDeploymentsAndTest(t, f, 1)
	assert.Equal(t, d, deployments[0])
}

func TestListDeploymentsWithStats(t *testing.T) {
	f := newDbFixture(true)         // no foreign keys
	listDeploymentsAndTest(t, f, 0) // ensure no deployments

	// create deployment and ensure there are no stats
	d := newDeployment(testBuildID, testEnvID)
	createDeploymentAndTest(t, f, d)

	dWS := scheme.DeploymentWithStats{Deployment: d} // no stats, all zeros
	deployments := listDeploymentsWithStatsAndTest(t, f, 1)
	assert.Equal(t, dWS, deployments[0])

	// add session for the stats
	s := newSession(f, sdktypes.CompletedSessionStateType)
	createSessionWithTest(t, f, s)

	// ensure that new session is included in stats
	dWS.Completed = 1
	deployments = listDeploymentsWithStatsAndTest(t, f, 1)
	assert.Equal(t, dWS, deployments[0])

	// delete session
	assert.NoError(t, f.gormdb.deleteSession(f.ctx, s.SessionID))
	testSessionDeleted(t, f, s.SessionID)

	// and check that deployment stats are updated
	deployments = listDeploymentsWithStatsAndTest(t, f, 1)
	dWS.Completed = 0 // completed session was deleted
	assert.Equal(t, dWS, deployments[0])
}

func TestDeleteDeployments(t *testing.T) {
	f := newDbFixture(true)         // no foreign keys
	listDeploymentsAndTest(t, f, 0) // ensure no deployments

	d := newDeployment(testBuildID, testEnvID)
	createDeploymentAndTest(t, f, d)

	// add sessions and check that deployment stats are updated
	session1 := newSession(f, sdktypes.CompletedSessionStateType)
	createSessionWithTest(t, f, session1)
	session2 := newSession(f, sdktypes.ErrorSessionStateType)
	createSessionWithTest(t, f, session2)

	dWS := scheme.DeploymentWithStats{Deployment: d, Completed: 1, Error: 1}
	deployments := listDeploymentsWithStatsAndTest(t, f, 1)
	assert.Equal(t, dWS, deployments[0])

	// delete deployment and check sessions are marked as deleted as well
	assert.NoError(t, f.gormdb.deleteDeployment(f.ctx, d.DeploymentID))

	listDeploymentsWithStatsAndTest(t, f, 0)
	testSessionDeleted(t, f, session1.SessionID)
	testSessionDeleted(t, f, session2.SessionID)
}

func TestForeignKeysDeployments(t *testing.T) {
	f := newDbFixture(false) // with foreign keys
	deployment := newDeployment("unexistingBuildID", "unexistingEnvID")
	err := f.gormdb.createDeployment(f.ctx, deployment)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "FOREIGN KEY")
}
