package dbgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

func (f *dbFixture) saveBuildsAndAssert(t *testing.T, builds ...scheme.Build) {
	for _, build := range builds {
		assert.NoError(t, f.gormdb.saveBuild(f.ctx, &build))
		findAndAssertOne(t, f, build, "build_id = ?", build.BuildID)
	}
}

func assertBuildDeleted(t *testing.T, f *dbFixture, buildID string) {
	assertSoftDeleted(t, f, scheme.Build{BuildID: buildID})
}

func preBuildTest(t *testing.T) *dbFixture {
	f := newDBFixture()
	findAndAssertCount(t, f, scheme.Build{}, 0, "")
	return f
}

func TestSaveBuild(t *testing.T) {
	f := preBuildTest(t)

	b := f.newBuild()
	// test saveBuid
	f.saveBuildsAndAssert(t, b)
}

func TestDeleteBuild(t *testing.T) {
	f := preBuildTest(t)

	b := f.newBuild()
	f.saveBuildsAndAssert(t, b)

	// test deleteBuild
	assert.NoError(t, f.gormdb.deleteBuild(f.ctx, b.BuildID))
	assertBuildDeleted(t, f, b.BuildID)
}

func TestListBuild(t *testing.T) {
	f := preBuildTest(t)

	// no builds
	flt := sdkservices.ListBuildsFilter{} // no Limit
	builds, err := f.gormdb.listBuilds(f.ctx, flt)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(builds))

	// create build and obtain it via list
	b := f.newBuild()
	f.saveBuildsAndAssert(t, b)

	// check listBuilds API
	builds, err = f.gormdb.listBuilds(f.ctx, flt)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(builds))
	assert.Equal(t, b, builds[0])

	// delete build
	assert.NoError(t, f.gormdb.deleteBuild(f.ctx, b.BuildID))

	// check listBuilds API - ensure no builds are found
	builds, err = f.gormdb.listBuilds(f.ctx, flt)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(builds))
}
