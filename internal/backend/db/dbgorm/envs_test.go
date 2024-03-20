package dbgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
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

func TestCreateEnv(t *testing.T) {
	f := newDBFixture(true)                       // no foreign keys
	findAndAssertCount(t, f, scheme.Env{}, 0, "") // no envs

	e := f.newEnv()
	// test createEnv
	f.createEnvsAndAssert(t, e)
}

func TestDeleteEnv(t *testing.T) {
	f := newDBFixture(true)                       // no foreign keys
	findAndAssertCount(t, f, scheme.Env{}, 0, "") // no envs

	e := f.newEnv()
	f.createEnvsAndAssert(t, e)

	// test deleteEnv
	assert.NoError(t, f.gormdb.deleteEnv(f.ctx, e.EnvID))
	f.assertEnvDeleted(t, e)
}

func TestDeleteEnvs(t *testing.T) {
	f := newDBFixture(true)                       // no foreign keys
	findAndAssertCount(t, f, scheme.Env{}, 0, "") // no envs

	e1, e2 := f.newEnv(), f.newEnv()
	f.createEnvsAndAssert(t, e1, e2)

	// test deleteEnvs
	assert.NoError(t, f.gormdb.deleteEnvs(f.ctx, []string{e1.EnvID, e2.EnvID}))
	f.assertEnvDeleted(t, e1, e2)
}

func TestDeleteEnvForeignKeys(t *testing.T) {
	f := newDBFixture(false)
	findAndAssertCount(t, f, scheme.Env{}, 0, "") // no envs

	b := f.newBuild()
	p := f.newProject()
	e := f.newEnv()
	e.ProjectID = p.ProjectID
	d := f.newDeployment()
	d.BuildID = b.BuildID
	d.EnvID = e.EnvID

	f.saveBuildsAndAssert(t, b)
	f.createProjectsAndAssert(t, p)
	f.createEnvsAndAssert(t, e)
	f.createDeploymentsAndAssert(t, d)

	// cannot delete env, since deployment referencing it
	err := f.gormdb.deleteEnv(f.ctx, e.EnvID)
	assert.ErrorContains(t, err, "FOREIGN KEY")

	// delete deployment (referencing build), then env
	assert.NoError(t, f.gormdb.deleteDeployment(f.ctx, d.DeploymentID))
	assert.NoError(t, f.gormdb.deleteEnv(f.ctx, e.EnvID))
	f.assertEnvDeleted(t, e)
}
