package dbgorm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/backend/internal/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func createDeploymentWithTest(t *testing.T, ctx context.Context, gormdb *gormdb, deployment scheme.Deployment) {
	assert.NoError(t, gormdb.createDeployment(ctx, deployment))
	res := gormdb.db.First(&scheme.Deployment{}, "deployment_id = ?", deployment.DeploymentID)
	assert.NoError(t, res.Error)
	assert.Equal(t, int64(1), res.RowsAffected)
}

func TestCreateDeployment(t *testing.T) {
	f := newDbFixture()

	deployment := makeSchemeDeployment()
	createDeploymentWithTest(t, f.ctx, f.gormdb, deployment)

	// obtain all deployments records from the deployments table
	var deployments []scheme.Deployment
	assert.NoError(t, f.db.Find(&deployments).Error)
	assert.Equal(t, int(1), len(deployments))
	assert.Equal(t, deployment, deployments[0])
}

func TestGetDeployment(t *testing.T) {
	f := newDbFixture()

	deployment := makeSchemeDeployment()
	createDeploymentWithTest(t, f.ctx, f.gormdb, deployment)

	d, err := f.gormdb.getDeployment(f.ctx, deployment.DeploymentID)
	assert.NoError(t, err)
	assert.Equal(t, d.DeploymentID, deployment.DeploymentID)

	// TODO: check after delete
}

func TestListDeployments(t *testing.T) {
	f := newDbFixture()

	flt := sdkservices.ListDeploymentsFilter{
		State:               sdktypes.DeploymentStateUnspecified,
		Limit:               0,
		IncludeSessionStats: false,
	}

	// no deployments
	deployments, err := f.gormdb.listDeployments(f.ctx, flt)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(deployments))

	// create deployment and obtain it via list
	deployment := makeSchemeDeployment()
	createDeploymentWithTest(t, f.ctx, f.gormdb, deployment)

	deployments, err = f.gormdb.listDeployments(f.ctx, flt)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(deployments))
	assert.Equal(t, deployment.DeploymentID, deployments[0].DeploymentID)
}

func TestListDeploymentsWithStats(t *testing.T) {
	f := newDbFixture()

	flt := sdkservices.ListDeploymentsFilter{
		State:               sdktypes.DeploymentStateUnspecified,
		Limit:               0,
		IncludeSessionStats: true,
	}

	// no deployments
	deployments, err := f.gormdb.listDeploymentsWithStats(f.ctx, flt)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(deployments))

	// create deployment and obtain it via list
	deployment := makeSchemeDeployment()
	createDeploymentWithTest(t, f.ctx, f.gormdb, deployment)

	deployments, err = f.gormdb.listDeploymentsWithStats(f.ctx, flt)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(deployments))
	assert.Equal(t, deployment.DeploymentID, deployments[0].DeploymentID)

	deploymentWS := scheme.DeploymentWithStats{Deployment: deployment}
	assert.Equal(t, deploymentWS, deployments[0])

	// add session for the stats
	session := makeSchemeSession()
	createSessionWithTest(t, f.ctx, f.gormdb, session)

	deployments, err = f.gormdb.listDeploymentsWithStats(f.ctx, flt)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(deployments))
	assert.Equal(t, deployment.DeploymentID, deployments[0].DeploymentID)

	deploymentWS.Completed = 1
	assert.Equal(t, deploymentWS, deployments[0])

	// selete session and check that deployment stats are updated
	assert.NoError(t, f.gormdb.deleteSession(f.ctx, session.SessionID))

	// session creation was tested in createSessionWithTest
	res := f.db.First(&scheme.Session{}, "session_id = ?", session.SessionID)
	assert.Equal(t, res.Error, gorm.ErrRecordNotFound)

	deployments, err = f.gormdb.listDeploymentsWithStats(f.ctx, flt)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(deployments))
	assert.Equal(t, deployment.DeploymentID, deployments[0].DeploymentID)
	deploymentWS.Completed = 0 // completed session was deleted
	assert.Equal(t, deploymentWS, deployments[0])
}
