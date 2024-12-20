package vars

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authz"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/secrets"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func TestMain(m *testing.M) {
	authz.DisableCheckForTesting()
	m.Run()
}

type secretsMock struct {
	secrets.Secrets
	SetFunc func(ctx context.Context, key string, value string) error
}

func (s secretsMock) Set(ctx context.Context, key string, value string) error {
	return s.SetFunc(ctx, key, value)
}

type dbMock struct {
	db.DB
	GetVarsFunc func(context.Context, sdktypes.VarScopeID, []sdktypes.Symbol) (sdktypes.Vars, error)
	SetVarsFunc func(context.Context, []sdktypes.Var) error
}

func (d dbMock) Transaction(ctx context.Context, f func(tx db.DB) error) error {
	return f(d)
}

func (d dbMock) AddOwnership(ctx context.Context, entities ...any) error {
	return nil
}

func (d dbMock) GetVars(ctx context.Context, sid sdktypes.VarScopeID, sym []sdktypes.Symbol) ([]sdktypes.Var, error) {
	return d.GetVarsFunc(ctx, sid, sym)
}

func (d dbMock) SetVars(ctx context.Context, vars []sdktypes.Var) error {
	return d.SetVarsFunc(ctx, vars)
}

func TestGetVarNotFound(t *testing.T) {
	d := dbMock{}
	callCounter := 0

	d.GetVarsFunc = func(ctx context.Context, vsi sdktypes.VarScopeID, s []sdktypes.Symbol) (sdktypes.Vars, error) {
		callCounter = callCounter + 1
		return nil, nil
	}

	v := Vars{
		db:      d,
		secrets: secretsMock{},
	}

	vv := sdktypes.NewVar(sdktypes.NewSymbol("efi")).SetValue("value")
	result, err := v.Get(context.Background(), vv.ScopeID(), sdktypes.Symbol{})

	require.Nil(t, err, "error should not be")
	require.Empty(t, result, "no vars should be returned")
	require.Equal(t, 1, callCounter, "GetVars was not called")
}

func TestGetVarFound(t *testing.T) {
	d := dbMock{}
	callCounter := 0
	vv := sdktypes.NewVar(sdktypes.NewSymbol("efi")).SetValue("value")

	d.GetVarsFunc = func(context.Context, sdktypes.VarScopeID, []sdktypes.Symbol) (sdktypes.Vars, error) {
		callCounter = callCounter + 1
		return []sdktypes.Var{vv}, nil
	}

	v := Vars{
		db:      d,
		secrets: secretsMock{},
	}

	result, err := v.Get(context.Background(), vv.ScopeID(), sdktypes.Symbol{})

	require.Nil(t, err, "error should not be")
	require.Len(t, result, 1, "should return one variable")
	require.Equal(t, 1, callCounter, "GetVars was not called")
	require.Equal(t, result[0], vv)
}

func TestSetVar(t *testing.T) {
	d := dbMock{}
	dbCallCounter := 0

	d.SetVarsFunc = func(ctx context.Context, v []sdktypes.Var) error {
		dbCallCounter = dbCallCounter + 1
		return nil
	}

	s := secretsMock{}
	sCallCounter := 0

	s.SetFunc = func(ctx context.Context, key string, value string) error {
		sCallCounter = sCallCounter + 1
		return nil
	}

	v := Vars{
		db:      d,
		secrets: secretsMock{},
	}

	va := sdktypes.NewVar(sdktypes.NewSymbol("test")).SetValue("value")
	err := v.Set(context.TODO(), va)

	require.Nil(t, err)
	require.Equal(t, sCallCounter, 0, "should not call secrets service set function")
	require.Equal(t, dbCallCounter, 1, "should call db set var")
}

func TestSetSecretVar(t *testing.T) {
	d := dbMock{}
	dbCallCounter := 0

	d.SetVarsFunc = func(ctx context.Context, v []sdktypes.Var) error {
		dbCallCounter = dbCallCounter + 1
		return nil
	}

	s := secretsMock{}
	sCallCounter := 0

	actualSecretKey := ""
	actualSecretValue := ""
	s.SetFunc = func(ctx context.Context, key string, value string) error {
		actualSecretKey = key
		actualSecretValue = value
		sCallCounter = sCallCounter + 1
		return nil
	}

	v := Vars{
		db:      d,
		secrets: s,
	}

	expectedValue := "value"
	va := sdktypes.NewVar(sdktypes.NewSymbol("test")).SetValue(expectedValue).SetSecret(true)
	expectedSecretKey := varSecretKey(va)
	err := v.Set(context.TODO(), va)

	require.Nil(t, err)
	require.Equal(t, sCallCounter, 1, "should call secrets service set function")
	require.Equal(t, dbCallCounter, 1, "should call db set var")
	require.Equal(t, actualSecretKey, expectedSecretKey)
	require.Equal(t, actualSecretValue, expectedValue)
}

func TestSetMultipleSecretVar(t *testing.T) {
	d := dbMock{}
	dbCallCounter := 0

	dbVals := []string{}
	d.SetVarsFunc = func(ctx context.Context, v []sdktypes.Var) error {
		dbVals = kittehs.Transform(v, func(v sdktypes.Var) string {
			return v.Value()
		})
		dbCallCounter = dbCallCounter + 1
		return nil
	}

	s := secretsMock{}
	sCallCounter := 0

	keys := []string{}
	vals := []string{}
	s.SetFunc = func(ctx context.Context, key string, value string) error {
		keys = append(keys, key)
		vals = append(vals, value)
		sCallCounter = sCallCounter + 1
		return nil
	}

	v := Vars{
		db:      d,
		secrets: s,
	}

	va := sdktypes.NewVar(sdktypes.NewSymbol("test")).SetValue("1").SetSecret(true)
	va2 := sdktypes.NewVar(sdktypes.NewSymbol("test2")).SetValue("2").SetSecret(true)

	err := v.Set(context.TODO(), va, va2)

	require.Nil(t, err)
	require.Equal(t, sCallCounter, 2, "should call secrets service set function")
	require.Equal(t, dbCallCounter, 1, "should call db set var")

	require.Equal(t, keys[0], varSecretKey(va))
	require.Equal(t, vals[0], va.Value())
	require.Equal(t, dbVals[0], keys[0])
	require.Equal(t, keys[1], varSecretKey(va2))
	require.Equal(t, vals[1], va2.Value())
	require.Equal(t, dbVals[1], keys[1])
}
