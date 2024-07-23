package dbgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func preOwnershipTestPermissive(t *testing.T) *dbFixture {
	f := preOwnershipTest(t)
	f.gormdb.owner = &PermissiveOwnershipChecker{f.gormdb.z}
	f.gormdb.z.Info("patched ownership checker", zap.String("type", "permissive"))
	return f
}

// Testing only builds, just for simplicity

func TestCreateBuildWithOwnershipP(t *testing.T) {
	f := preOwnershipTestPermissive(t)

	p := f.newProject()
	f.createProjectsAndAssert(t, p)

	b := f.newBuild(p)
	f.saveBuildsAndAssert(t, b)
	assert.NoError(t, f.gormdb.isCtxUserEntity(f.ctx, b.BuildID))

	// different user - authorized to create build for the project owned by different user, thus failing with duplicate key
	f.withUser(u2)
	assert.ErrorIs(t, f.gormdb.saveBuild(f.ctx, &b), gorm.ErrDuplicatedKey)
}

func TestDeleteBuildsWithOwnershipP(t *testing.T) {
	f := preOwnershipTestPermissive(t)

	b := f.newBuild()
	f.saveBuildsAndAssert(t, b)

	// different user - authorized to delete build owned by different user
	f.withUser(u2)
	assert.NoError(t, f.gormdb.deleteBuild(f.ctx, b.BuildID))
}

func TestGetBuildWithOwnershipP(t *testing.T) {
	f := preOwnershipTestPermissive(t)

	b := f.newBuild()
	f.saveBuildsAndAssert(t, b)

	// different user - authorized to get build owned by different user
	f.withUser(u2)
	_, err := f.gormdb.getBuild(f.ctx, b.BuildID)
	assert.NoError(t, err)
}

func TestListBuildsWithOwnershipP(t *testing.T) {
	f := preOwnershipTestPermissive(t)

	b := f.newBuild()
	f.saveBuildsAndAssert(t, b)

	// different user
	f.withUser(u2)
	builds, err := f.gormdb.listBuilds(f.ctx, sdkservices.ListBuildsFilter{})
	assert.Len(t, builds, 1) // allowed to fetch builds owned by other users
	assert.NoError(t, err)
}

func TestSetVarWithOwnershipP(t *testing.T) {
	f := preOwnershipTestPermissive(t)

	c, env := createConnectionAndEnv(t, f)
	v1 := f.newVar("k", "v", env) // env scoped var
	v2 := f.newVar("k", "v", c)   // connection scoped var

	// different user - authorised to create var for the connection and env owned by different user
	f.withUser(u2)
	f.setVarsAndAssert(t, v1)
	f.setVarsAndAssert(t, v2)
}
