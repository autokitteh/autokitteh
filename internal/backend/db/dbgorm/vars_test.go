package dbgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
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
	vars[0].VarID = sdktypes.InvalidVarID.UUIDValue()
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
