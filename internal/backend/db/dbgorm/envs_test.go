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
	f := newDBFixture().withUser(sdktypes.DefaultUser)
	findAndAssertCount[scheme.Env](t, f, 0, "") // no envs
	return f
}

func createProjectAndEnv(t *testing.T, f *dbFixture) (scheme.Project, scheme.Env) {
	p := f.newProject()
	e := f.newEnv(p)

	f.createProjectsAndAssert(t, p)
	f.createEnvsAndAssert(t, e)

	return p, e
}

func TestCreateEnv(t *testing.T) {
	f := preEnvTest(t)

	// test createEnv
	_, _ = createProjectAndEnv(t, f)
}

func TestGetEnv(t *testing.T) {
	f := preEnvTest(t)

	p, e := createProjectAndEnv(t, f)

	// test getEnvByName
	e2, err := f.gormdb.getEnvByName(f.ctx, p.ProjectID, e.Name)
	assert.NoError(t, err)
	assert.Equal(t, e, *e2)

	// test getEnvByID
	e2, err = f.gormdb.getEnvByID(f.ctx, e.EnvID)
	assert.NoError(t, err)
	assert.Equal(t, e, *e2)

	// delete env
	assert.NoError(t, f.gormdb.deleteEnv(f.ctx, e.EnvID))

	// test getEnvByName after delete
	_, err = f.gormdb.getEnvByName(f.ctx, p.ProjectID, e.Name)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	// test getEnvByID after delete
	_, err = f.gormdb.getEnvByID(f.ctx, e.EnvID)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestListEnvs(t *testing.T) {
	f := preEnvTest(t)

	p, e := createProjectAndEnv(t, f)

	envs, err := f.gormdb.listEnvs(f.ctx, p.ProjectID)
	assert.NoError(t, err)
	assert.Len(t, envs, 1)
	assert.Equal(t, e, envs[0])

	// delete env
	assert.NoError(t, f.gormdb.deleteEnv(f.ctx, e.EnvID))

	// test listEnvs after delete
	envs, err = f.gormdb.listEnvs(f.ctx, p.ProjectID)
	assert.NoError(t, err)
	assert.Len(t, envs, 0)
}

func TestCreateEnvForeignKeys(t *testing.T) {
	f := preEnvTest(t)

	p := f.newProject()
	b := f.newBuild()
	e := f.newEnv()
	f.createProjectsAndAssert(t, p)
	f.saveBuildsAndAssert(t, b)

	// negative tests with non-existing assets
	// zero ProjectID
	assert.Equal(t, e.ProjectID, sdktypes.UUID{}) // zero value
	assert.ErrorIs(t, f.gormdb.createEnv(f.ctx, &e), gorm.ErrForeignKeyViolated)

	// use existing user-owner buildID to fake unexisting projectID
	e.ProjectID = b.BuildID // no such projectID, since it's a buildID
	assert.ErrorIs(t, f.gormdb.createEnv(f.ctx, &e), gorm.ErrForeignKeyViolated)

	e.ProjectID = p.ProjectID
	f.createEnvsAndAssert(t, e)
}

func TestDeleteEnv(t *testing.T) {
	f := preEnvTest(t)

	_, e := createProjectAndEnv(t, f)

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
