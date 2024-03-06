package dbgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

func saveBuildAndAssert(t *testing.T, f *dbFixture, build scheme.Build) {
	assert.NoError(t, f.gormdb.saveBuild(f.ctx, build))
	findAndAssertOne(t, f, build, "build_id = ?", build.BuildID)
}

func TestSaveBuild(t *testing.T) {
	f := newDbFixture(true)
	build := newBuild()
	saveBuildAndAssert(t, f, build)
	findAndAssertOne(t, f, build, "") // check there is only single build in DB
}

func TestDeleteBuild(t *testing.T) {
	f := newDbFixture(true)
	build := newBuild()
	saveBuildAndAssert(t, f, build)

	// delete build and ensure it's completely deleted (use Unscoped to check it's not just soft deleted)
	assert.NoError(t, f.gormdb.deleteBuild(f.ctx, build.BuildID))
	res := f.db.Unscoped().First(&scheme.Build{}, "build_id = ?", build.BuildID)
	assert.ErrorAs(t, gorm.ErrRecordNotFound, &res.Error)
}

func TestListBuild(t *testing.T) {
	f := newDbFixture(true)
	flt := sdkservices.ListBuildsFilter{} // no Limit

	// no builds
	builds, err := f.gormdb.listBuilds(f.ctx, flt)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(builds))

	// create build and obtain it via list
	build := newBuild()
	saveBuildAndAssert(t, f, build)

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
