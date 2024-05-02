package dbgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
)

func (f *dbFixture) setVarsAndAssert(t *testing.T, vars ...scheme.Var) {
	assert.NoError(t, f.gormdb.setVars(f.ctx, vars))
	for _, vr := range vars {
		findAndAssertOne(t, f, vr, "scope_id = ? and name = ?", vr.ScopeID, vr.Name)
	}
}

func (f *dbFixture) assertVarDeleted(t *testing.T, vars ...scheme.Var) {
	for _, vr := range vars {
		assertDeleted(t, f, vr)
	}
}

func preVarTest(t *testing.T) *dbFixture {
	f := newDBFixture()
	findAndAssertCount(t, f, scheme.Var{}, 0, "") // no vars
	return f
}

func TestSetVar(t *testing.T) {
	f := preVarTest(t)

	v := f.newVar("foo", "bar")
	// test setVar
	f.setVarsAndAssert(t, v)

	// test modify
	v.Value = "baz"
	f.setVarsAndAssert(t, v)
}

func TestGetVar(t *testing.T) {
	f := preVarTest(t)

	v := f.newVar("foo", "bar")
	f.setVarsAndAssert(t, v)

	// test getVar
	vars, err := f.gormdb.getVars(f.ctx, v.ScopeID, v.Name)
	assert.NoError(t, err)
	assert.Equal(t, v, vars[0])
}

func TestDeleteVar(t *testing.T) {
	f := preVarTest(t)

	v := f.newVar("foo", "bar")
	f.setVarsAndAssert(t, v)

	// test deleteEnvVar
	assert.NoError(t, f.gormdb.deleteVars(f.ctx, v.ScopeID, v.Name))
	f.assertVarDeleted(t, v)
}

/*
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
	findAndAssertCount(t, f, scheme.Env{}, 0, "") // no envs
	return f
}

func TestCreateEnv(t *testing.T) {
	f := preEnvTest(t)

	e := f.newEnv()
	// test createEnv
	f.createEnvsAndAssert(t, e)
}

func TestCreateEnvForeignKeys(t *testing.T) {
	f := preEnvTest(t)

	e := f.newEnv()
	unexisting := scheme.UUIDOrNil(sdktypes.NewProjectID().UUIDValue())

	e.ProjectID = unexisting
	assert.ErrorIs(t, f.gormdb.createEnv(f.ctx, &e), gorm.ErrForeignKeyViolated)

	p := f.newProject()
	e.ProjectID = &p.ProjectID
	f.createProjectsAndAssert(t, p)
	f.createEnvsAndAssert(t, e)
}

func TestDeleteEnv(t *testing.T) {
	f := preEnvTest(t)

	e := f.newEnv()
	f.createEnvsAndAssert(t, e)

	// test deleteEnv
	assert.NoError(t, f.gormdb.deleteEnv(f.ctx, e.EnvID))
	f.assertEnvDeleted(t, e)
}

func TestDeleteEnvs(t *testing.T) {
	f := preEnvTest(t)

	e1, e2 := f.newEnv(), f.newEnv()
	f.createEnvsAndAssert(t, e1, e2)

	// test deleteEnvs
	assert.NoError(t, f.gormdb.deleteEnvs(f.ctx, []sdktypes.UUID{e1.EnvID, e2.EnvID}))
	f.assertEnvDeleted(t, e1, e2)
}

func TestDeleteEnvForeignKeys(t *testing.T) {
	f := preEnvTest(t)

	b := f.newBuild()
	p := f.newProject()
	e := f.newEnv()
	e.ProjectID = &p.ProjectID
	d := f.newDeployment()
	d.BuildID = &b.BuildID
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
*/
