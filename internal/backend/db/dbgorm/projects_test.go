package dbgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
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
		assertSoftDeleted(t, f, project)
	}
}

func preProjectTest(t *testing.T) *dbFixture {
	f := newDBFixture().withUser(sdktypes.DefaultUser)
	findAndAssertCount[scheme.Project](t, f, 0, "") // no projects
	return f
}

func TestCreateProject(t *testing.T) {
	f := preProjectTest(t)

	p := f.newProject()
	// test createProject
	f.createProjectsAndAssert(t, p)
}

func TestCreateDuplicatedProjectName(t *testing.T) {
	f := preProjectTest(t)

	p := f.newProject()
	// test createProject
	f.createProjectsAndAssert(t, p)

	// create different project with the same name
	p2 := f.newProject()
	p2.Name = p.Name
	assert.Equal(t, p.Name, p2.Name)
	assert.NotEqual(t, p.ProjectID, p2.ProjectID)

	// test create another project with the same name
	assert.ErrorIs(t, f.gormdb.createProject(f.ctx, &p2), gorm.ErrDuplicatedKey)

	// delete project
	assert.NoError(t, f.gormdb.deleteProject(f.ctx, p.ProjectID))
	f.assertProjectDeleted(t, p)

	// test create another project with the same name after delete
	f.createProjectsAndAssert(t, p2)
}

func TestGetProjects(t *testing.T) {
	f := preProjectTest(t)

	p := f.newProject()
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
	_, err = f.gormdb.getProjectByName(f.ctx, p.Name)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestListProjects(t *testing.T) {
	f := preProjectTest(t)

	p := f.newProject()
	f.createProjectsAndAssert(t, p)

	// test listProjects
	projects := f.listProjectsAndAssert(t, 1)
	assert.Equal(t, p, projects[0])

	// test listProjects after delete
	assert.NoError(t, f.gormdb.deleteProject(f.ctx, p.ProjectID))
	f.listProjectsAndAssert(t, 0)
}

func TestGetProjectDeployments(t *testing.T) {
	f := preProjectTest(t)

	// create 4 envs. E1 with 3 deployments, E2 with 1 (dupl with E1) and E3 with 0.
	// p1:
	// - e1: (d1, d2)
	// - e2: (d3)
	// - e3
	//
	// p2:
	// - e4: (d4)
	p1, p2 := f.newProject(), f.newProject()
	e1, e2, e3, e4 := f.newEnv(), f.newEnv(), f.newEnv(), f.newEnv()
	d1, d2, d3, d4 := f.newDeployment(), f.newDeployment(), f.newDeployment(), f.newDeployment()
	b := f.newBuild()

	e1.ProjectID = p1.ProjectID
	e2.ProjectID = p1.ProjectID
	e3.ProjectID = p1.ProjectID
	e4.ProjectID = p2.ProjectID

	d1.EnvID = &e1.EnvID
	d2.EnvID = &e1.EnvID
	d3.EnvID = &e2.EnvID
	d4.EnvID = &e4.EnvID

	d1.BuildID = b.BuildID
	d2.BuildID = b.BuildID
	d3.BuildID = b.BuildID
	d4.BuildID = b.BuildID

	f.createProjectsAndAssert(t, p1, p2)
	f.createEnvsAndAssert(t, e1, e2, e3, e4)
	f.saveBuildsAndAssert(t, b)
	f.createDeploymentsAndAssert(t, d1, d2, d3, d4)

	// ds, err := f.gormdb.getProjectDeployments(f.ctx, p1.ProjectID)
	// assert.NoError(t, err)
	// assert.Equal(t, []sdktypes.UUID{d1.DeploymentID, d2.DeploymentID, d3.DeploymentID},
	// 	kittehs.Transform(ds, func(d DeploymentState) sdktypes.UUID { return d.DeploymentID }))
}

func TestGetProjectEnvs(t *testing.T) {
	f := preProjectTest(t)

	// create two envs - one with deployment and second without
	p := f.newProject()
	e1, e2 := f.newEnv(), f.newEnv()
	d := f.newDeployment()
	b := f.newBuild()

	e1.ProjectID = p.ProjectID
	e2.ProjectID = p.ProjectID
	d.EnvID = &e1.EnvID
	d.BuildID = b.BuildID

	f.createProjectsAndAssert(t, p)
	f.createEnvsAndAssert(t, e1, e2)
	f.saveBuildsAndAssert(t, b)
	f.createDeploymentsAndAssert(t, d)

	envIDs, err := f.gormdb.getProjectEnvs(f.ctx, p.ProjectID)
	assert.NoError(t, err)
	// ensure that we got both envs - even of there is no deployments attached
	assert.Equal(t, []sdktypes.UUID{e1.EnvID, e2.EnvID}, envIDs)
}

func TestDeleteProjectAndDependents(t *testing.T) {
	f := preProjectTest(t)

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
	p1, p2 := f.newProject(), f.newProject()

	c := f.newConnection()
	c.IntegrationID = &testIntegrationID
	c.ProjectID = &p1.ProjectID

	e1p1, e2p1, e1p2 := f.newEnv(), f.newEnv(), f.newEnv()
	e1p1.ProjectID = p1.ProjectID
	e2p1.ProjectID = p1.ProjectID
	e1p2.ProjectID = p2.ProjectID

	t1, t2 := f.newTrigger(), f.newTrigger()
	t1.ProjectID = p1.ProjectID
	t1.EnvID = e1p1.EnvID
	t1.ConnectionID = c.ConnectionID
	t2.ProjectID = p1.ProjectID
	t2.EnvID = e2p1.EnvID
	t2.ConnectionID = c.ConnectionID

	sig := f.newSignal()
	sig.ConnectionID = c.ConnectionID

	b := f.newBuild()

	d1e1p1, d2e1p1, d1e2p1, d1e1p2 := f.newDeployment(), f.newDeployment(), f.newDeployment(), f.newDeployment()
	d1e1p1.EnvID = &e1p1.EnvID
	d2e1p1.EnvID = &e1p1.EnvID
	d1e2p1.EnvID = &e2p1.EnvID
	d1e1p2.EnvID = &e1p2.EnvID
	d1e1p1.BuildID = b.BuildID
	d2e1p1.BuildID = b.BuildID
	d1e2p1.BuildID = b.BuildID
	d1e1p2.BuildID = b.BuildID

	s1d1e1p1 := f.newSession(sdktypes.SessionStateTypeCompleted)
	s2d1e2p1 := f.newSession(sdktypes.SessionStateTypeError)
	s3d1e1p2 := f.newSession(sdktypes.SessionStateTypeCompleted)
	s1d1e1p1.DeploymentID = &d1e1p1.DeploymentID
	s2d1e2p1.DeploymentID = &d1e2p1.DeploymentID
	s3d1e1p2.DeploymentID = &d1e1p2.DeploymentID

	f.createProjectsAndAssert(t, p1, p2)
	f.createConnectionsAndAssert(t, c)
	f.createEnvsAndAssert(t, e1p1, e2p1, e1p2)
	f.createTriggersAndAssert(t, t1, t2)
	f.saveSignalsAndAssert(t, sig)
	f.saveBuildsAndAssert(t, b)
	f.createDeploymentsAndAssert(t, d1e1p1, d2e1p1, d1e2p1, d1e1p2)
	f.createSessionsAndAssert(t, s1d1e1p1, s2d1e2p1, s3d1e1p2)

	// ensure failure if deployments are not inactive
	err := f.gormdb.deleteProjectAndDependents(f.ctx, p1.ProjectID)
	assert.ErrorIs(t, err, sdkerrors.ErrFailedPrecondition)

	// set deployments state to inactive and delete project and its dependents.
	for _, d := range []*scheme.Deployment{&d1e1p1, &d2e1p1, &d1e2p1} {
		_, err := f.gormdb.updateDeploymentState(f.ctx, d.DeploymentID, sdktypes.DeploymentStateInactive)
		assert.NoError(t, err)
	}

	err = f.gormdb.deleteProjectAndDependents(f.ctx, p1.ProjectID)
	assert.NoError(t, err)

	f.assertDeploymentsDeleted(t, d1e1p1, d2e1p1, d1e2p1)
	f.assertSessionsDeleted(t, s1d1e1p1, s2d1e2p1)
	f.assertEnvDeleted(t, e1p1, e2p1)
	f.assertTriggersDeleted(t, t1, t2)
	f.assertSignalsDeleted(t, sig)
	f.assertConnectionDeleted(t, c)
	f.assertProjectDeleted(t, p1)
}

func TestUpdateProject(t *testing.T) {
	f := preProjectTest(t).WithDebug()

	p := f.newProject()
	f.createProjectsAndAssert(t, p)

	// update project
	p.Name = p.Name + "_updated"
	assert.NoError(t, f.gormdb.updateProject(f.ctx, &p))
	findAndAssertOne(t, f, p, "project_id = ?", p.ProjectID)
}
