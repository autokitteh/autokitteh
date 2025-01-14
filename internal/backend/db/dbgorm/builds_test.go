package dbgorm

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

func (f *dbFixture) saveBuildsAndAssert(t *testing.T, builds ...scheme.Build) {
	for _, build := range builds {
		assert.NoError(t, f.gormdb.saveBuild(f.ctx, &build))
		findAndAssertOne(t, f, build, "build_id = ?", build.BuildID)
	}
}

func (f *dbFixture) assertBuildDeleted(t *testing.T, buildID uuid.UUID) {
	assertDeleted(t, f, scheme.Build{BuildID: buildID})
}

func preBuildTest(t *testing.T) (*dbFixture, scheme.Project) {
	f := newDBFixture()
	findAndAssertCount[scheme.Build](t, f, 0, "")

	p := f.newProject()
	f.createProjectsAndAssert(t, p)

	return f, p
}

func TestSaveBuild(t *testing.T) {
	f, p := preBuildTest(t)

	b := f.newBuild(p)

	// test saveBuild
	f.saveBuildsAndAssert(t, b)
}

func TestDeleteBuild(t *testing.T) {
	f, p := preBuildTest(t)

	b := f.newBuild(p)
	f.saveBuildsAndAssert(t, b)

	// test deleteBuild
	assert.NoError(t, f.gormdb.deleteBuild(f.ctx, b.BuildID))
	f.assertBuildDeleted(t, b.BuildID)
}

func TestDeleteProjectDeleteBuild(t *testing.T) {
	f, p := preBuildTest(t)
	b := f.newBuild(p)
	f.saveBuildsAndAssert(t, b)

	// test deleteProject
	assert.NoError(t, f.gormdb.deleteProject(f.ctx, p.ProjectID))
	f.assertBuildDeleted(t, b.BuildID)
}

func TestDeleteBuildDeleteDeployment(t *testing.T) {
	f, p := preBuildTest(t)
	b := f.newBuild(p)
	f.saveBuildsAndAssert(t, b)

	d := f.newDeployment(b, p)
	f.createDeploymentsAndAssert(t, d)

	// test deleteBuild
	assert.NoError(t, f.gormdb.deleteBuild(f.ctx, b.BuildID))
	f.assertDeploymentsDeleted(t, d)
}

func TestDeleteBuildDeleteSession(t *testing.T) {
	f, p := preBuildTest(t)
	b := f.newBuild(p)
	f.saveBuildsAndAssert(t, b)

	s := f.newSession(b, p)
	f.createSessionsAndAssert(t, s)

	// test deleteBuild
	assert.NoError(t, f.gormdb.deleteBuild(f.ctx, b.BuildID))
	f.assertSessionsDeleted(t, s)
}

func TestGetBuild(t *testing.T) {
	f, p := preBuildTest(t)

	b := f.newBuild(p)
	f.saveBuildsAndAssert(t, b)

	// test getBuild
	b2, err := f.gormdb.getBuild(f.ctx, b.BuildID)
	assert.NoError(t, err)
	assert.Equal(t, b, *b2)

	// test getBuild after delete
	assert.NoError(t, f.gormdb.deleteBuild(f.ctx, b.BuildID))
	_, err = f.gormdb.getBuild(f.ctx, b.BuildID)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestListBuilds(t *testing.T) {
	f, p := preBuildTest(t)

	// no builds
	flt := sdkservices.ListBuildsFilter{} // no Limit
	builds, err := f.gormdb.listBuilds(f.ctx, flt)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(builds))

	// create build and obtain it via list
	b := f.newBuild(p)
	f.saveBuildsAndAssert(t, b)

	// check listBuilds API
	builds, err = f.gormdb.listBuilds(f.ctx, flt)
	assert.NoError(t, err)
	assert.Len(t, builds, 1)
	assert.Equal(t, b, builds[0])

	// delete build
	assert.NoError(t, f.gormdb.deleteBuild(f.ctx, b.BuildID))

	// check listBuilds API - ensure no builds are found
	builds, err = f.gormdb.listBuilds(f.ctx, flt)
	assert.NoError(t, err)
	assert.Len(t, builds, 0)
}
