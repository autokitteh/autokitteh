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

	e := f.newEnv()
	// test createEnv
	f.WithForeignKeysDisabled(func() { f.createEnvsAndAssert(t, e) })
}

func TestCreateEnvForeignKeys(t *testing.T) {
	f := preEnvTest(t)

	e := f.newEnv()
	unexisting := sdktypes.NewProjectID().UUIDValue()

	e.ProjectID = unexisting
	assert.ErrorIs(t, f.gormdb.createEnv(f.ctx, &e), gorm.ErrForeignKeyViolated)

	p := f.newProject()
	e.ProjectID = p.ProjectID
	f.createProjectsAndAssert(t, p)
	f.createEnvsAndAssert(t, e)
}

func TestDeleteEnv(t *testing.T) {
	f := preEnvTest(t)

	e := f.newEnv()
	f.WithForeignKeysDisabled(func() { f.createEnvsAndAssert(t, e) })

	// test deleteEnv
	assert.NoError(t, f.gormdb.deleteEnv(f.ctx, e.EnvID))
	f.assertEnvDeleted(t, e)
}

func TestDeleteEnvs(t *testing.T) {
	f := preEnvTest(t)

	e1, e2 := f.newEnv(), f.newEnv()
	f.WithForeignKeysDisabled(func() { f.createEnvsAndAssert(t, e1, e2) })

	// test deleteEnvs
	assert.NoError(t, f.gormdb.deleteEnvs(f.ctx, []sdktypes.UUID{e1.EnvID, e2.EnvID}))
	f.assertEnvDeleted(t, e1, e2)
}

func TestDeleteEnvForeignKeys(t *testing.T) {
	f := preEnvTest(t)

	b := f.newBuild()
	p := f.newProject()
	e := f.newEnv()
	e.ProjectID = p.ProjectID
	d := f.newDeployment()
	d.BuildID = b.BuildID
	d.EnvID = &e.EnvID

	f.saveBuildsAndAssert(t, b)
	f.createProjectsAndAssert(t, p)
	f.createEnvsAndAssert(t, e)
	f.createDeploymentsAndAssert(t, d)

	// cannot delete env, since deployment referencing it
	assert.ErrorIs(t, f.gormdb.deleteEnv(f.ctx, e.EnvID), gorm.ErrForeignKeyViolated)

	// delete deployment (referencing build), then env
	assert.NoError(t, f.gormdb.deleteDeployment(f.ctx, d.DeploymentID))
	assert.NoError(t, f.gormdb.deleteEnv(f.ctx, e.EnvID))
	f.assertEnvDeleted(t, e)
}
