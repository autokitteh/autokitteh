package dbgorm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/backend/internal/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

func testSaveBuild(t *testing.T, ctx context.Context, gormdb *gormdb, build scheme.Build) {
	assert.Nil(t, gormdb.saveBuild(ctx, build, []byte{}))
	testDBExistsOne(t, gormdb, build, "build_id = ?", build.BuildID)
}

func TestSaveBuild(t *testing.T) {
	f := newDbFixture()

	build := makeSchemeBuild()
	testSaveBuild(t, f.ctx, f.gormdb, build)

	testDBExistsOne(t, f.gormdb, build, "") // check there is only single build in DB
}

func TestDeleteBuild(t *testing.T) {
	f := newDbFixture()

	build := makeSchemeBuild()
	testSaveBuild(t, f.ctx, f.gormdb, build)

	// delete
	assert.Nil(t, f.gormdb.deleteBuild(f.ctx, build.BuildID))

	// ensure that build is completely deleted (use Unscoped to check that it is not just soft-deleted)
	res := f.db.Unscoped().First(&scheme.Build{}, "build_id = ?", build.BuildID)
	assert.Equal(t, res.Error, gorm.ErrRecordNotFound)
}

func TestListBuild(t *testing.T) {
	f := newDbFixture()

	flt := sdkservices.ListBuildsFilter{} // no Limit

	// no builds
	builds, err := f.gormdb.listBuilds(f.ctx, flt)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(builds))

	// create build and obtain it via list
	build := makeSchemeBuild()
	testSaveBuild(t, f.ctx, f.gormdb, build)

	builds, err = f.gormdb.listBuilds(f.ctx, flt)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(builds))
	assert.Equal(t, build, builds[0])

	// delete and ensure that list is empty
	assert.Nil(t, f.gormdb.deleteBuild(f.ctx, build.BuildID))

	builds, err = f.gormdb.listBuilds(f.ctx, flt)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(builds))
}
