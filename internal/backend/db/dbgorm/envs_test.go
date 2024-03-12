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

func assertEnvDeleted(t *testing.T, f *dbFixture, envs ...scheme.Env) {
	for _, env := range envs {
		assertSoftDeleted(t, f, scheme.Env{EnvID: env.EnvID})
	}
}

func TestCreateEnv(t *testing.T) {
	f := newDBFixture(true)                       // no foreign keys
	findAndAssertCount(t, f, scheme.Env{}, 0, "") // no envs

	e := newEnv(f)
	// test createEnv
	f.createEnvsAndAssert(t, e)
}

func TestDeleteEnv(t *testing.T) {
	f := newDBFixture(true)                       // no foreign keys
	findAndAssertCount(t, f, scheme.Env{}, 0, "") // no envs

	e := newEnv(f)
	f.createEnvsAndAssert(t, e)

	// test deleteEnv
	assert.NoError(t, f.gormdb.deleteEnv(f.ctx, e.EnvID))
	assertEnvDeleted(t, f, e)
}

func TestDeleteEnvs(t *testing.T) {
	f := newDBFixture(true)                       // no foreign keys
	findAndAssertCount(t, f, scheme.Env{}, 0, "") // no envs

	e1, e2 := newEnv(f), newEnv(f)
	f.createEnvsAndAssert(t, e1, e2)

	// test deleteEnvs
	assert.NoError(t, f.gormdb.deleteEnvs(f.ctx, []string{e1.EnvID, e2.EnvID}))
	assertEnvDeleted(t, f, e1, e2)
}

func TestDeleteEnvForeignKeys(t *testing.T) {
	f := newDBFixture(false)
	findAndAssertCount(t, f, scheme.Env{}, 0, "") // no envs

	b := newBuild()
	p := newProject(f)
	e := newEnv(f)
	e.ProjectID = p.ProjectID
	d := newDeployment(f)
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
	assertEnvDeleted(t, f, e)
}
