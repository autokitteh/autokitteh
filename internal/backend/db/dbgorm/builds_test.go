package dbgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

func saveBuildsAndAssert(t *testing.T, f *dbFixture, builds ...scheme.Build) {
	for _, build := range builds {
		assert.NoError(t, f.gormdb.saveBuild(f.ctx, &build))
		findAndAssertOne(t, f, build, "build_id = ?", build.BuildID)
	}
}

func assertBuildDeleted(t *testing.T, f *dbFixture, buildID string) {
	assertSoftDeleted(t, f, scheme.Build{BuildID: buildID})
}

func TestSaveBuild(t *testing.T) {
	f := newDBFixture(true)
	build := newBuild()
	saveBuildsAndAssert(t, f, build)
	findAndAssertOne(t, f, build, "") // check there is only single build in DB
}

func TestDeleteBuild(t *testing.T) {
	f := newDBFixture(true)
	build := newBuild()
	saveBuildsAndAssert(t, f, build)

	assert.NoError(t, f.gormdb.deleteBuild(f.ctx, build.BuildID))
	assertBuildDeleted(t, f, build.BuildID)
}

func TestListBuild(t *testing.T) {
	f := newDBFixture(true)
	flt := sdkservices.ListBuildsFilter{} // no Limit

	// no builds
	builds, err := f.gormdb.listBuilds(f.ctx, flt)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(builds))

	// create build and obtain it via list
	build := newBuild()
	saveBuildsAndAssert(t, f, build)

	// check listBuilds API
	builds, err = f.gormdb.listBuilds(f.ctx, flt)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(builds))
	assert.Equal(t, build, builds[0])

	// delete build
	assert.NoError(t, f.gormdb.deleteBuild(f.ctx, build.BuildID))

	// check listBuilds API - ensure no builds are found
	builds, err = f.gormdb.listBuilds(f.ctx, flt)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(builds))
}
