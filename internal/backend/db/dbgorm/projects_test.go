package dbgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func createProjectsAndAssert(t *testing.T, f *dbFixture, projects ...scheme.Project) {
	for _, project := range projects {
		assert.NoError(t, f.gormdb.createProject(f.ctx, project))
		findAndAssertOne(t, f, project, "project_id = ?", project.ProjectID)
	}
}

func listProjectsAndAssert(t *testing.T, f *dbFixture, expected int) []scheme.Project {
	projects, err := f.gormdb.listProjects(f.ctx)
	assert.NoError(t, err)
	assert.Equal(t, expected, len(projects))
	return projects
}

func assertProjectDeleted(t *testing.T, f *dbFixture, projectID string) {
	assertSoftDeleted(t, f, scheme.Project{ProjectID: projectID})
}

func TestCreateProject(t *testing.T) {
	f := newDbFixture(true)                           // no foreign keys
	findAndAssertCount(t, f, scheme.Project{}, 0, "") // no projects

	p := newProject(f)
	// test createProject
	createProjectsAndAssert(t, f, p)
}

func TestGetProjects(t *testing.T) {
	f := newDbFixture(true)                           // no foreign keys
	findAndAssertCount(t, f, scheme.Project{}, 0, "") // no projects

	p := newProject(f)
	createProjectsAndAssert(t, f, p)

	// test getProjectByID
	project, err := f.gormdb.getProject(f.ctx, p.ProjectID)
	assert.NoError(t, err)
	assert.Equal(t, p, *project)

	// test getProjectByName
	project, err = f.gormdb.getProjectByName(f.ctx, p.Name)
	assert.NoError(t, err)
	assert.Equal(t, p, *project)

	// delete project
	assert.NoError(t, f.gormdb.deleteProject(f.ctx, p.ProjectID))

	// test getProjectByName after delete
	_, err = f.gormdb.getProject(f.ctx, p.ProjectID)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	// test getProjectByName after delete
	_, err = f.gormdb.getProject(f.ctx, p.Name)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestListProjects(t *testing.T) {
	f := newDbFixture(true)                           // no foreign keys
	findAndAssertCount(t, f, scheme.Project{}, 0, "") // no projects

	p := newProject(f)
	createProjectsAndAssert(t, f, p)

	// test listProjects
	projects := listProjectsAndAssert(t, f, 1)
	assert.Equal(t, p, projects[0])

	// test listProjects after delete
	assert.NoError(t, f.gormdb.deleteProject(f.ctx, p.ProjectID))
	listProjectsAndAssert(t, f, 0)
}

func TestDeleteProject(t *testing.T) {
	f := newDbFixture(true)                           // no foreign keys
	findAndAssertCount(t, f, scheme.Project{}, 0, "") // no projects

	p := newProject(f)
	createProjectsAndAssert(t, f, p)

	// delete project
	assert.NoError(t, f.gormdb.deleteProject(f.ctx, p.ProjectID))
	findAndAssertCount(t, f, scheme.Project{}, 0, "") // no projects
	assertProjectDeleted(t, f, p.ProjectID)
}

	findAndAssertCount(t, f, scheme.Project{}, 0, "") // no projects



}


func TestDeleteProjectAndDependents(t *testing.T) {
	f := newDbFixture(true)                           // no foreign keys
	findAndAssertCount(t, f, scheme.Project{}, 0, "") // no projects

	// initialize:
	// - p1
	//   - e1
	//     - d1 (s1), d2
	//   - e2
	//     - d1 (s2)
	// - p2
	//   - e1
	//     - d1 (s3)
	p1, p2 := newProject(f), newProject(f)
	createProjectsAndAssert(t, f, p1)
	createProjectsAndAssert(t, f, p2)

	e1p1, e2p1, e1p2 := newEnv(f), newEnv(f), newEnv(f)
	e1p1.ProjectID = p1.ProjectID
	e2p1.ProjectID = p1.ProjectID
	e1p2.ProjectID = p2.ProjectID
	createEnvsAndAssert(t, f, e1p1)
	createEnvsAndAssert(t, f, e2p1)
	createEnvsAndAssert(t, f, e1p2)

	d1e1p1 := newDeployment(f)
	d2e1p1 := newDeployment(f)
	d1e2p1 := newDeployment(f)
	d1e1p2 := newDeployment(f)
	d1e1p1.EnvID = e1p1.EnvID
	d2e1p1.EnvID = e1p1.EnvID
	d1e2p1.EnvID = e2p1.EnvID
	d1e1p2.EnvID = e1p2.EnvID
	createDeploymentsAndAssert(t, f, d1e1p1)
	createDeploymentsAndAssert(t, f, d2e1p1)
	createDeploymentsAndAssert(t, f, d1e2p1)
	createDeploymentsAndAssert(t, f, d1e1p2)

	s1d1e1p1 := newSession(f, sdktypes.SessionStateTypeCompleted)
	s2d1e2p1 := newSession(f, sdktypes.SessionStateTypeError)
	s3d1e1p2 := newSession(f, sdktypes.SessionStateTypeCompleted)
	s1d1e1p1.DeploymentID = d1e1p1.DeploymentID
	s2d1e2p1.DeploymentID = d1e2p1.DeploymentID
	s3d1e1p2.DeploymentID = d1e1p2.DeploymentID
	createSessionAndAssert(t, f, s1d1e1p1)
	createSessionAndAssert(t, f, s2d1e2p1)
	createSessionAndAssert(t, f, s3d1e1p2)

	// ensure failure if deployments are not inactive
	err := f.gormdb.deleteProjectAndDependents(f.ctx, p1.ProjectID)
	assert.ErrorIs(t, err, sdkerrors.ErrFailedPrecondition)

	// set deployments state to inactive and delete project and its dependents.
	for _, d := range []*scheme.Deployment{&d1e1p1, &d2e1p1, &d1e2p1} {
		assert.NoError(t, f.gormdb.updateDeploymentState(f.ctx, d.DeploymentID, sdktypes.DeploymentStateInactive))
	}

	err = f.gormdb.deleteProjectAndDependents(f.ctx, p1.ProjectID)
	assert.NoError(t, err)

	assertDeploymentDeleted(t, f, d1e1p1.DeploymentID)
	assertDeploymentDeleted(t, f, d2e1p1.DeploymentID)
	assertDeploymentDeleted(t, f, d1e2p1.DeploymentID)
	assertEnvDeleted(t, f, e1p1.EnvID)
	assertEnvDeleted(t, f, e2p1.EnvID)
	assertProjectDeleted(t, f, p1.ProjectID)

	assertSessionDeleted(t, f, s1d1e1p1.SessionID)
	assertSessionDeleted(t, f, s2d1e2p1.SessionID)
}
