package dbgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
)

func createEnvAndAssert(t *testing.T, f *dbFixture, env scheme.Env) {
	assert.NoError(t, f.gormdb.createEnv(f.ctx, env))
	findAndAssertOne(t, f, env, "env_id = ?", env.EnvID)
}

func assertEnvDeleted(t *testing.T, f *dbFixture, envID string) {
	assertSoftDeleted(t, f, scheme.Env{EnvID: envID})
}

func TestCreateEnv(t *testing.T) {
	f := newDbFixture(true)                       // no foreign keys
	findAndAssertCount(t, f, scheme.Env{}, 0, "") // no envs

	e := newEnv(f)
	// test createEnv
	createEnvAndAssert(t, f, e)
}

func TestDeleteEnv(t *testing.T) {
	f := newDbFixture(true)                       // no foreign keys
	findAndAssertCount(t, f, scheme.Env{}, 0, "") // no envs

	e := newEnv(f)
	createEnvAndAssert(t, f, e)

	// test deleteEnv
	assert.NoError(t, f.gormdb.deleteEnv(f.ctx, e.EnvID))
	findAndAssertCount(t, f, scheme.Env{}, 0, "") // no envs
}

func TestDeleteEnvForeignKeys(t *testing.T) {
	f := newDbFixture(false)
	findAndAssertCount(t, f, scheme.Env{}, 0, "") // no envs

	b := newBuild()
	p := newProject(f)
	e := newEnv(f)
	e.ProjectID = p.ProjectID
	d := newDeploymentWithBuildAndEnv(f, b, e)

	saveBuildAndAssert(t, f, b)
	createProjectAndAssert(t, f, p)
	createEnvAndAssert(t, f, e)
	createDeploymentAndAssert(t, f, d)

	// cannot delete env, since deployment referencing it
	err := f.gormdb.deleteEnv(f.ctx, e.EnvID)
	assert.ErrorContains(t, err, "FOREIGN KEY")

	// delete deployment (referencing build), then build , then env
	assert.NoError(t, f.gormdb.deleteDeployment(f.ctx, d.DeploymentID))
	assert.NoError(t, f.gormdb.deleteBuild(f.ctx, b.BuildID))
	assert.NoError(t, f.gormdb.deleteEnv(f.ctx, e.EnvID))
	assertEnvDeleted(t, f, e.EnvID)
}
