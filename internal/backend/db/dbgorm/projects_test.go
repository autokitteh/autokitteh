package dbgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (f *dbFixture) createProjectsAndAssert(t *testing.T, projects ...scheme.Project) {
	for _, project := range projects {
		assert.NoError(t, f.gormdb.createProject(f.ctx, &project))
		findAndAssertOne(t, f, project, "project_id = ?", project.ProjectID)
	}
}

func (f *dbFixture) listProjectsAndAssert(t *testing.T, expected int) []scheme.Project {
	projects, err := f.gormdb.listProjects(f.ctx)
	assert.NoError(t, err)
	assert.Equal(t, expected, len(projects))
	return projects
}

func (f *dbFixture) assertProjectDeleted(t *testing.T, projects ...scheme.Project) {
	for _, project := range projects {
		assertSoftDeleted(t, f, scheme.Project{ProjectID: project.ProjectID})
	}
}

func TestCreateProject(t *testing.T) {
	f := newDBFixture(true)                           // no foreign keys
	findAndAssertCount(t, f, scheme.Project{}, 0, "") // no projects

	p := newProject(f)
	// test createProject
	f.createProjectsAndAssert(t, p)
}

func TestGetProjects(t *testing.T) {
	f := newDBFixture(true)                           // no foreign keys
	findAndAssertCount(t, f, scheme.Project{}, 0, "") // no projects

	p := newProject(f)
	f.createProjectsAndAssert(t, p)

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
	f := newDBFixture(true)                           // no foreign keys
	findAndAssertCount(t, f, scheme.Project{}, 0, "") // no projects

	p := newProject(f)
	f.createProjectsAndAssert(t, p)

	// test listProjects
	projects := f.listProjectsAndAssert(t, 1)
	assert.Equal(t, p, projects[0])

	// test listProjects after delete
	assert.NoError(t, f.gormdb.deleteProject(f.ctx, p.ProjectID))
	f.listProjectsAndAssert(t, 0)
}

func TestGetProjectDeployments(t *testing.T) {
	f := newDBFixture(true)                           // no foreign keys
	findAndAssertCount(t, f, scheme.Project{}, 0, "") // no projects

	// create 4 envs. E1 with 3 deployments, E2 with 1 (dupl with E1) and E3 with 0.
	// p1:
	// - e1: (d1, d2)
	// - e2: (d3)
	// - e3
	//
	// p2:
	// - e4: (d4)
	p1, p2 := newProject(f), newProject(f)
	e1, e2, e3, e4 := newEnv(f), newEnv(f), newEnv(f), newEnv(f)
	d1, d2, d3, d4 := newDeployment(f), newDeployment(f), newDeployment(f), newDeployment(f)

	e1.ProjectID = p1.ProjectID
	e2.ProjectID = p1.ProjectID
	e3.ProjectID = p1.ProjectID
	e4.ProjectID = p2.ProjectID

	d1.EnvID = e1.EnvID
	d2.EnvID = e1.EnvID
	d3.EnvID = e2.EnvID
	d4.EnvID = e4.EnvID

	f.createProjectsAndAssert(t, p1, p2)
	f.createEnvsAndAssert(t, e1, e2, e3, e4)
	f.createDeploymentsAndAssert(t, d1, d2, d3, d4)

	ds, err := f.gormdb.getProjectDeployments(f.ctx, p1.ProjectID)
	assert.NoError(t, err)
	assert.Equal(t, []string{d1.DeploymentID, d2.DeploymentID, d3.DeploymentID},
		kittehs.Transform(ds, func(d DeploymentState) string { return d.DeploymentID }))
}

func TestGetProjectEnvs(t *testing.T) {
	f := newDBFixture(true)                           // no foreign keys
	findAndAssertCount(t, f, scheme.Project{}, 0, "") // no projects

	// create two envs - one with deployment and second without
	p := newProject(f)
	e1, e2 := newEnv(f), newEnv(f)
	d := newDeployment(f)

	e1.ProjectID = p.ProjectID
	e2.ProjectID = p.ProjectID
	d.EnvID = e1.EnvID

	f.createProjectsAndAssert(t, p)
	f.createEnvsAndAssert(t, e1, e2)
	f.createDeploymentsAndAssert(t, d)

	envIDs, err := f.gormdb.getProjectEnvs(f.ctx, p.ProjectID)
	assert.NoError(t, err)
	// ensure that we got both envs - even of there is no deployments attached
	assert.Equal(t, []string{e1.EnvID, e2.EnvID}, envIDs)
}

func TestDeleteProjectAndDependents(t *testing.T) {
	f := newDBFixture(false)
	findAndAssertCount(t, f, scheme.Project{}, 0, "") // no projects

	// initialize:
	// - p1
	//   - e1
	//     - d1 (s1), d2
	//   - e2
	//     - d1 (s2)
	//   - t1, t2
	// - p2
	//   - e1
	//     - d1 (s3)
	p1, p2 := newProject(f), newProject(f)

	i := newIntegration()
	c := newConnection()
	c.IntegrationID = i.IntegrationID
	c.ProjectID = p1.ProjectID

	e1p1, e2p1, e1p2 := newEnv(f), newEnv(f), newEnv(f)
	e1p1.ProjectID = p1.ProjectID
	e2p1.ProjectID = p1.ProjectID
	e1p2.ProjectID = p2.ProjectID

	t1, t2 := newTrigger(f), newTrigger(f)
	t1.ProjectID = p1.ProjectID
	t1.EnvID = e1p1.EnvID
	t1.ConnectionID = c.ConnectionID
	t2.ProjectID = p1.ProjectID
	t2.EnvID = e2p1.EnvID
	t1.ConnectionID = c.ConnectionID

	b := newBuild()

	d1e1p1 := newDeployment(f)
	d2e1p1 := newDeployment(f)
	d1e2p1 := newDeployment(f)
	d1e1p2 := newDeployment(f)
	d1e1p1.EnvID = e1p1.EnvID
	d2e1p1.EnvID = e1p1.EnvID
	d1e2p1.EnvID = e2p1.EnvID
	d1e1p2.EnvID = e1p2.EnvID
	d1e1p1.BuildID = b.BuildID
	d2e1p1.BuildID = b.BuildID
	d1e2p1.BuildID = b.BuildID
	d1e1p2.BuildID = b.BuildID

	s1d1e1p1 := newSession(f, sdktypes.SessionStateTypeCompleted)
	s2d1e2p1 := newSession(f, sdktypes.SessionStateTypeError)
	s3d1e1p2 := newSession(f, sdktypes.SessionStateTypeCompleted)
	s1d1e1p1.DeploymentID = d1e1p1.DeploymentID
	s2d1e2p1.DeploymentID = d1e2p1.DeploymentID
	s3d1e1p2.DeploymentID = d1e1p2.DeploymentID

	f.createProjectsAndAssert(t, p1, p2)
	f.createIntegrationsAndAssert(t, i)
	f.createConnectionsAndAssert(t, c)
	f.createEnvsAndAssert(t, e1p1, e2p1, e1p2)
	f.createTriggersAndAssert(t, t1, t2)
	f.saveBuildsAndAssert(t, b)
	f.createDeploymentsAndAssert(t, d1e1p1, d2e1p1, d1e2p1, d1e1p2)
	f.createSessionsAndAssert(t, s1d1e1p1, s2d1e2p1, s3d1e1p2)

	// ensure failure if deployments are not inactive
	err := f.gormdb.deleteProjectAndDependents(f.ctx, p1.ProjectID)
	assert.ErrorIs(t, err, sdkerrors.ErrFailedPrecondition)

	// set deployments state to inactive and delete project and its dependents.
	for _, d := range []*scheme.Deployment{&d1e1p1, &d2e1p1, &d1e2p1} {
		assert.NoError(t, f.gormdb.updateDeploymentState(f.ctx, d.DeploymentID, sdktypes.DeploymentStateInactive))
	}

	err = f.gormdb.deleteProjectAndDependents(f.ctx, p1.ProjectID)
	assert.NoError(t, err)

	f.assertDeploymentsDeleted(t, d1e1p1, d2e1p1, d1e2p1)
	f.assertEnvDeleted(t, e1p1, e2p1)
	f.assertProjectDeleted(t, p1)
	f.assertSessionsDeleted(t, s1d1e1p1, s2d1e2p1)
	f.assertTriggersDeleted(t, t1, t2)
}
