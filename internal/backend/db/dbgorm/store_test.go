package dbgorm_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func TestValues(t *testing.T) {
	pids := []sdktypes.ProjectID{sdktypes.NewProjectID(), sdktypes.NewProjectID(), sdktypes.NewProjectID()}

	db, err := dbgorm.New(zap.NewNop(), &dbgorm.Config{})
	require.NoError(t, err)

	ctx := t.Context()

	require.NoError(t, db.Connect(ctx))
	require.NoError(t, db.Setup(ctx))

	oid, err := db.CreateOrg(ctx, sdktypes.NewOrg().WithID(sdktypes.NewOrgID()))
	require.NoError(t, err)

	require.NoError(t, db.CreateProject(ctx, sdktypes.NewProject().WithName(sdktypes.NewSymbol("test1")).WithID(pids[0]).WithOrgID(oid)))
	require.NoError(t, db.CreateProject(ctx, sdktypes.NewProject().WithName(sdktypes.NewSymbol("test2")).WithID(pids[1]).WithOrgID(oid)))
	require.NoError(t, db.CreateProject(ctx, sdktypes.NewProject().WithName(sdktypes.NewSymbol("test3")).WithID(pids[2]).WithOrgID(oid)))

	require.NoError(t, db.SetStoreValue(ctx, pids[0], "key0", sdktypes.NewIntegerValue(10)))
	require.NoError(t, db.SetStoreValue(ctx, pids[0], "key1", sdktypes.NewIntegerValue(11)))
	require.NoError(t, db.SetStoreValue(ctx, pids[1], "key", sdktypes.NewIntegerValue(2)))

	vs, err := db.ListStoreValues(ctx, pids[0], nil, true)
	if assert.NoError(t, err) {
		assert.Equal(t, map[string]sdktypes.Value{
			"key0": sdktypes.NewIntegerValue(10),
			"key1": sdktypes.NewIntegerValue(11),
		}, vs)
	}

	vs, err = db.ListStoreValues(ctx, pids[1], nil, true)
	if assert.NoError(t, err) {
		assert.Equal(t, map[string]sdktypes.Value{"key": sdktypes.NewIntegerValue(2)}, vs)
	}

	vs, err = db.ListStoreValues(ctx, pids[2], nil, true)
	if assert.NoError(t, err) {
		assert.Equal(t, map[string]sdktypes.Value{}, vs)
	}

	require.NoError(t, db.SetStoreValue(ctx, pids[0], "key0", sdktypes.NewIntegerValue(100)))

	vs, err = db.ListStoreValues(ctx, pids[0], nil, true)
	if assert.NoError(t, err) {
		assert.Equal(t, map[string]sdktypes.Value{
			"key0": sdktypes.NewIntegerValue(100),
			"key1": sdktypes.NewIntegerValue(11),
		}, vs)
	}

	vs, err = db.ListStoreValues(ctx, pids[0], nil, false)
	if assert.NoError(t, err) {
		assert.Equal(t, map[string]sdktypes.Value{
			"key0": sdktypes.InvalidValue,
			"key1": sdktypes.InvalidValue,
		}, vs)
	}

	vs, err = db.ListStoreValues(ctx, pids[0], []string{"key1"}, false)
	if assert.NoError(t, err) {
		assert.Equal(t, map[string]sdktypes.Value{
			"key1": sdktypes.InvalidValue,
		}, vs)
	}

	vs, err = db.ListStoreValues(ctx, pids[0], []string{"key666"}, false)
	if assert.NoError(t, err) {
		assert.Equal(t, map[string]sdktypes.Value{}, vs)
	}

	v, err := db.GetStoreValue(ctx, pids[0], "key0")
	if assert.NoError(t, err) {
		assert.Equal(t, sdktypes.NewIntegerValue(100), v)
	}

	v, err = db.GetStoreValue(ctx, pids[1], "key")
	if assert.NoError(t, err) {
		assert.Equal(t, sdktypes.NewIntegerValue(2), v)
	}

	v, err = db.GetStoreValue(ctx, pids[2], "key")
	assert.NoError(t, err)
	assert.Equal(t, sdktypes.Nothing, v)
}
