package dbgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// specialied version of findAndAssertOne in order to exclude var.ID from the comparison
func findAndAssertOneVar(t *testing.T, f *dbFixture, v scheme.Var) {
	vr := findAndAssertCount[scheme.Var](t, f, 1, "scope_id = ? and name = ?", v.ScopeID, v.Name)[0]
	// clear ID for comparision
	vr.VarID = sdktypes.InvalidVarID.UUIDValue()
	v.VarID = vr.VarID
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

func preVarTest(t *testing.T) *dbFixture {
	f := newDBFixture()
	findAndAssertCount[scheme.Var](t, f, 0, "") // no vars

	p := f.newProject() // parent project
	f.createProjectsAndAssert(t, p)

	c := f.newConnection()
	f.createConnectionsAndAssert(t, c)

	e := f.newEnv()
	e.ProjectID = p.ProjectID
	f.createEnvsAndAssert(t, e)

	f.projectID = p.ProjectID
	f.connectionID = c.ConnectionID
	f.envID = e.EnvID

	varIDfunc = func() sdktypes.UUID { return newTestID() }

	return f
}

func TestSetVar(t *testing.T) {
	f := preVarTest(t)

	// test setVar
	// scopeID isn't set to eother connectioID or envID, thus not in user scope
	v1 := f.newVar("v1", "connectionScope")
	assert.ErrorIs(t, f.gormdb.setVar(f.ctx, &v1), sdkerrors.ErrUnauthorized)
	v1.ScopeID = f.connectionID
	f.setVarsAndAssert(t, v1)

	v2 := f.newVar("v2", "envScope")
	assert.ErrorIs(t, f.gormdb.setVar(f.ctx, &v2), sdkerrors.ErrUnauthorized)
	v2.ScopeID = f.envID
	f.setVarsAndAssert(t, v2)

	// remove env (soft delete) and test that var cannot be added (foreign key emulation)
	assert.NoError(t, f.gormdb.deleteEnv(f.ctx, f.envID))
	assert.ErrorIs(t, f.gormdb.setVar(f.ctx, &v2), gorm.ErrForeignKeyViolated)
}

func TestReSetVar(t *testing.T) {
	f := preVarTest(t)

	v := f.newVar("foo", "bar")
	v.ScopeID = f.envID
	// test setVar
	f.setVarsAndAssert(t, v)

	// test modify
	v.Value = "baz"
	v.ScopeID = f.envID
	f.setVarsAndAssert(t, v)
}

func TestGetVar(t *testing.T) {
	f := preVarTest(t)

	v := f.newVar("foo", "bar")
	v.ScopeID = f.envID
	f.setVarsAndAssert(t, v)

	// test getVar
	vars, err := f.gormdb.getVars(f.ctx, v.ScopeID, v.Name)
	assert.NoError(t, err)
	vars[0].VarID = sdktypes.InvalidVarID.UUIDValue()
	assert.Equal(t, v, vars[0])
}

func TestDeleteVar(t *testing.T) {
	f := preVarTest(t)

	v := f.newVar("foo", "bar")
	v.ScopeID = f.envID
	f.setVarsAndAssert(t, v)

	// test deleteEnvVar
	assert.NoError(t, f.gormdb.deleteVars(f.ctx, v.ScopeID, v.Name))
	f.assertVarDeleted(t, v)
}
