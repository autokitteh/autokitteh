package dbgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (f *dbFixture) createEnvsAndAssert(t *testing.T, envs ...scheme.Env) {
	for _, env := range envs {
		assert.NoError(t, f.gormdb.createEnv(f.ctx, &env))
		findAndAssertOne(t, f, env, "env_id = ?", env.EnvID)
	}
}

func (f *dbFixture) assertEnvDeleted(t *testing.T, envs ...scheme.Env) {
	for _, env := range envs {
		assertSoftDeleted(t, f, env)
	}
}

func preEnvTest(t *testing.T) *dbFixture {
	f := newDBFixture()
	findAndAssertCount[scheme.Env](t, f, 0, "") // no envs
	return f
}

func TestCreateEnv(t *testing.T) {
	f := preEnvTest(t)

	p := f.newProject()
	f.createProjectsAndAssert(t, p)
	e := f.newEnv(p)

	// test createEnv
	f.createEnvsAndAssert(t, e)
}

func TestGetEnv(t *testing.T) {
	f := preEnvTest(t)

	p := f.newProject()
	e := f.newEnv(p)
	f.createProjectsAndAssert(t, p)
	f.createEnvsAndAssert(t, e)

	// test getEnvByName
	e2, err := f.gormdb.getEnvByName(f.ctx, p.ProjectID, e.Name)
	assert.NoError(t, err)
	assert.Equal(t, e, *e2)

	// test getEnvByID
	e2, err = f.gormdb.getEnvByID(f.ctx, e.EnvID)
	assert.NoError(t, err)
	assert.Equal(t, e, *e2)
}

func TestCreateEnvForeignKeys(t *testing.T) {
	f := preEnvTest(t)

	p := f.newProject()
	b := f.newBuild()
	e := f.newEnv()
	f.createProjectsAndAssert(t, p)
	f.saveBuildsAndAssert(t, b)

	// use user owner buildID to pass user checks as unexisting projectID
	e.ProjectID = b.BuildID // no such projectID, since it's a buildID
	assert.ErrorIs(t, f.gormdb.createEnv(f.ctx, &e), gorm.ErrForeignKeyViolated)

	e.ProjectID = p.ProjectID
	f.createEnvsAndAssert(t, e)
}

func TestDeleteEnv(t *testing.T) {
	f := preEnvTest(t)

	p := f.newProject()
	e := f.newEnv(p)
	f.createProjectsAndAssert(t, p)
	f.createEnvsAndAssert(t, e)

	// test deleteEnv
	assert.NoError(t, f.gormdb.deleteEnv(f.ctx, e.EnvID))
	f.assertEnvDeleted(t, e)
}

func TestDeleteEnvs(t *testing.T) {
	f := preEnvTest(t)

	p := f.newProject()
	e1, e2 := f.newEnv(p), f.newEnv(p)
	f.createProjectsAndAssert(t, p)
	f.createEnvsAndAssert(t, e1, e2)

	// test deleteEnvs
	assert.NoError(t, f.gormdb.deleteEnvs(f.ctx, []sdktypes.UUID{e1.EnvID, e2.EnvID}))
	f.assertEnvDeleted(t, e1, e2)
}

func TestDeleteEnvForeignKeys(t *testing.T) {
	f := preEnvTest(t)

	p := f.newProject()
	b := f.newBuild(p)
	e := f.newEnv(p)
	d := f.newDeployment(b, e)

	f.createProjectsAndAssert(t, p)
	f.saveBuildsAndAssert(t, b)
	f.createEnvsAndAssert(t, e)
	f.createDeploymentsAndAssert(t, d)

	// cannot delete env, since deployment referencing it
	assert.ErrorIs(t, f.gormdb.deleteEnv(f.ctx, e.EnvID), gorm.ErrForeignKeyViolated)

	// delete deployment (referencing build), then env
	assert.NoError(t, f.gormdb.deleteDeployment(f.ctx, d.DeploymentID))
	assert.NoError(t, f.gormdb.deleteEnv(f.ctx, e.EnvID))
	f.assertEnvDeleted(t, e)
}
