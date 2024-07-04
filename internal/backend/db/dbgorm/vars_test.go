package dbgorm

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// specialied version of findAndAssertOne in order to exclude var.ID from the comparison
func findAndAssertOneVar(t *testing.T, f *dbFixture, v scheme.Var) {
	vr := findAndAssertCount[scheme.Var](t, f, 1, "var_id = ? and name = ?", v.ScopeID, v.Name)[0]
	require.Equal(t, v, vr)
}

func (f *dbFixture) setVarsAndAssert(t *testing.T, vars ...scheme.Var) {
	for _, vr := range vars {
		assert.NoError(t, f.gormdb.setVar(f.ctx, &vr))
		findAndAssertOneVar(t, f, vr)
	}
}

func (f *dbFixture) assertVarDeleted(t *testing.T, vars ...scheme.Var) {
	for _, vr := range vars {
		assertDeleted(t, f, vr)
	}
}

func createConnectionAndEnv(t *testing.T, f *dbFixture) (scheme.Connection, scheme.Env) {
	p := f.newProject()
	c := f.newConnection(p)
	env := f.newEnv(p)

	f.createProjectsAndAssert(t, p)
	f.createConnectionsAndAssert(t, c)
	f.createEnvsAndAssert(t, env)

	return c, env
}

func preVarTest(t *testing.T) *dbFixture {
	f := newDBFixture()
	findAndAssertCount[scheme.Var](t, f, 0, "") // no vars
	return f
}

func TestSetVar(t *testing.T) {
	f := preVarTest(t)
	c, env := createConnectionAndEnv(t, f)

	// test setVar
	v1 := f.newVar("v1", "connectionScope", c)
	f.setVarsAndAssert(t, v1)

	v2 := f.newVar("v2", "envScope", env)
	f.setVarsAndAssert(t, v2)

	fmt.Println("!!! here")
	// test scopeID as foreign keys to either connectionID or envID
	assert.NoError(t, f.gormdb.deleteConnection(f.ctx, c.ConnectionID))
	assert.ErrorIs(t, f.gormdb.setVar(f.ctx, &v1), gorm.ErrForeignKeyViolated)

	assert.NoError(t, f.gormdb.deleteEnv(f.ctx, env.EnvID))
	assert.ErrorIs(t, f.gormdb.setVar(f.ctx, &v2), gorm.ErrForeignKeyViolated)

	// scopeID is zero, thus violates foreign key constraint
	v3 := f.newVar("v3", "invalid")
	assert.Equal(t, v3.ScopeID, sdktypes.UUID{}) // zero
	assert.Equal(t, v3.VarID, sdktypes.UUID{})   // zero
	assert.ErrorIs(t, f.gormdb.setVar(f.ctx, &v3), gorm.ErrForeignKeyViolated)
}

func TestReSetVar(t *testing.T) {
	f := preVarTest(t)
	_, env := createConnectionAndEnv(t, f)

	v := f.newVar("foo", "bar", env)
	// test setVar
	f.setVarsAndAssert(t, v)

	// test modify
	v.Value = "baz"
	f.setVarsAndAssert(t, v)
}

func TestListVarUnexisting(t *testing.T) {
	f := preVarTest(t)
	c := f.newConnection()
	f.createConnectionsAndAssert(t, c)

	v := f.newVar("k", "v", c)
	f.setVarsAndAssert(t, v)

	// test listVars
	vars, err := f.gormdb.listVars(f.ctx, c.ConnectionID, v.Name) // the same as v.ScopeID or v.VarID
	assert.NoError(t, err)
	assert.Len(t, vars, 1)
	assert.Equal(t, v, vars[0])

	// test listVars with non-existing var
	vars, err = f.gormdb.listVars(f.ctx, c.ConnectionID, "unexisting")
	assert.NoError(t, err)
	assert.Len(t, vars, 0)
}

func (f *dbFixture) testListVar(t *testing.T, v scheme.Var) {
	// test listVars
	vars, err := f.gormdb.listVars(f.ctx, v.ScopeID, v.Name)
	assert.NoError(t, err)
	assert.Len(t, vars, 1)
	assert.Equal(t, v, vars[0])

	// test listVars with non-existing var
	vars, err = f.gormdb.listVars(f.ctx, v.ScopeID, "unexisting")
	assert.NoError(t, err)
	assert.Len(t, vars, 0)

	// delete scope - either Connection or Env
	var o scheme.Ownership
	assert.NoError(t, f.db.Where("entity_id = ?", v.ScopeID).First(&o).Error)
	switch o.EntityType {
	case "Connection":
		assert.NoError(t, f.gormdb.deleteConnection(f.ctx, v.ScopeID))
	case "Env":
		assert.NoError(t, f.gormdb.deleteEnv(f.ctx, v.ScopeID))
	}

	// test var was deleted due to scope deletion
	f.assertVarDeleted(t, v)

	// test listVars after scope deletion
	vars, err = f.gormdb.listVars(f.ctx, v.ScopeID, v.Name)
	assert.NoError(t, err)
	assert.Len(t, vars, 0)
}

func TestListVar(t *testing.T) {
	f := preVarTest(t).WithDebug()

	c, e := createConnectionAndEnv(t, f)

	ve := f.newVar("k", "env scope", e)
	f.setVarsAndAssert(t, ve)
	f.testListVar(t, ve)

	vc := f.newVar("k", "connection scope", c)
	f.setVarsAndAssert(t, vc)
	f.testListVar(t, vc)
}

func TestDeleteVar(t *testing.T) {
	f := preVarTest(t)
	_, env := createConnectionAndEnv(t, f)

	v := f.newVar("foo", "bar", env)
	f.setVarsAndAssert(t, v)

	// test deleteEnvVar
	assert.NoError(t, f.gormdb.deleteVars(f.ctx, v.ScopeID, v.Name))
	f.assertVarDeleted(t, v)
}
