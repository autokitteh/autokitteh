package dbgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	projects, err := f.gormdb.listProjects(f.ctx, sdktypes.InvalidOrgID)
	assert.NoError(t, err)
	assert.Equal(t, expected, len(projects))
	return projects
}

func (f *dbFixture) assertProjectDeleted(t *testing.T, projects ...scheme.Project) {
	for _, project := range projects {
		assertDeleted(t, f, project)
	}
}

func preProjectTest(t *testing.T) *dbFixture {
	f := newDBFixture()
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
	resetTimes(project)
	assert.Equal(t, p, *project)

	// test getProjectByName
	project, err = f.gormdb.getProjectByName(f.ctx, sdktypes.InvalidOrgID, p.Name)
	assert.NoError(t, err)
	resetTimes(project)
	assert.Equal(t, p, *project)

	// delete project
	assert.NoError(t, f.gormdb.deleteProject(f.ctx, p.ProjectID))

	// test getProjectByName after delete
	_, err = f.gormdb.getProject(f.ctx, p.ProjectID)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	// test getProjectByName after delete
	_, err = f.gormdb.getProjectByName(f.ctx, sdktypes.InvalidOrgID, p.Name)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestListProjects(t *testing.T) {
	f := preProjectTest(t)

	p := f.newProject()
	f.createProjectsAndAssert(t, p)

	// test listProjects
	projects := f.listProjectsAndAssert(t, 1)
	resetTimes(&projects[0])
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
	d1, d2, d3, d4 := f.newDeployment(p1), f.newDeployment(p1), f.newDeployment(p1), f.newDeployment(p1)
	b := f.newBuild(p1)

	d1.BuildID = b.BuildID
	d2.BuildID = b.BuildID
	d3.BuildID = b.BuildID
	d4.BuildID = b.BuildID

	f.createProjectsAndAssert(t, p1, p2)
	f.saveBuildsAndAssert(t, b)
	f.createDeploymentsAndAssert(t, d1, d2, d3, d4)

	// ds, err := f.gormdb.getProjectDeployments(f.ctx, p1.ProjectID)
	// assert.NoError(t, err)
	// assert.Equal(t, []sdktypes.UUID{d1.DeploymentID, d2.DeploymentID, d3.DeploymentID},
	// 	kittehs.Transform(ds, func(d DeploymentState) sdktypes.UUID { return d.DeploymentID }))
}

func TestDeleteProjectAndDependents(t *testing.T) {
	f := preProjectTest(t)

	// initialize:
	// - p1
	//   - e1
	//     - d1 (s1)
	//     - d2
	//   - e2
	//     - d1 (s2)
	//   - t1, t2
	// - p2
	//   - e1
	//     - d1 (s3)

	p1, p2 := f.newProject(), f.newProject()
	c := f.newConnection(p1)

	t1, t2 := f.newTrigger(p1, c), f.newTrigger(p1, c)
	b := f.newBuild(p1)
	d1e1p1, d2e1p1, d1e2p1, d1e1p2 := f.newDeployment(p1, b), f.newDeployment(p1, b), f.newDeployment(p1, b), f.newDeployment(p2, b)

	s1d1e1p1 := f.newSession(sdktypes.SessionStateTypeCompleted, d1e1p1, p1, b)
	s2d1e2p1 := f.newSession(sdktypes.SessionStateTypeError, d1e2p1, p1, b)
	s3d1e1p2 := f.newSession(sdktypes.SessionStateTypeCompleted, d1e1p2, p2, b)

	sig := f.newSignal(c)

	f.createProjectsAndAssert(t, p1, p2)
	f.createConnectionsAndAssert(t, c)
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
	f.assertTriggersDeleted(t, t1, t2)
	f.assertSignalsDeleted(t, sig)
	f.assertConnectionDeleted(t, c)
	f.assertProjectDeleted(t, p1)
}

func TestUpdateProject(t *testing.T) {
	f := preProjectTest(t)

	p := f.newProject()
	f.createProjectsAndAssert(t, p)

	// update project
	p.Name += "_updated"
	assert.NoError(t, f.gormdb.updateProject(f.ctx, &p))
	res := findAndAssertCount[scheme.Project](t, f, 1, "project_id = ?", p.ProjectID)
	resetTimes(&res[0])
	require.Equal(t, p, res[0])
}
