package dbgorm_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var pids = []sdktypes.ProjectID{
	sdktypes.NewProjectID(),
	sdktypes.NewProjectID(),
	sdktypes.NewProjectID(),
}

func TestValues(t *testing.T) {
	db, err := dbgorm.New(zap.NewNop(), &dbgorm.Config{})
	require.NoError(t, err)

	ctx := context.Background()
	ctx = authcontext.SetAuthnUserID(ctx, sdktypes.BuiltinDefaultUserID.String())

	require.NoError(t, db.Connect(ctx))
	require.NoError(t, db.Setup(ctx))
	require.NoError(t, db.CreateProject(ctx, sdktypes.NewProject().WithName(sdktypes.NewSymbol("test1")).WithID(pids[0])))
	require.NoError(t, db.CreateProject(ctx, sdktypes.NewProject().WithName(sdktypes.NewSymbol("test2")).WithID(pids[1])))
	require.NoError(t, db.CreateProject(ctx, sdktypes.NewProject().WithName(sdktypes.NewSymbol("test3")).WithID(pids[2])))

	require.NoError(t, db.SetValue(ctx, pids[0], "key0", sdktypes.NewIntegerValue(10)))
	require.NoError(t, db.SetValue(ctx, pids[0], "key1", sdktypes.NewIntegerValue(11)))
	require.NoError(t, db.SetValue(ctx, pids[1], "key", sdktypes.NewIntegerValue(2)))

	vs, err := db.ListValues(ctx, pids[0])
	if assert.NoError(t, err) {
		assert.Equal(t, map[string]sdktypes.Value{
			"key0": sdktypes.NewIntegerValue(10),
			"key1": sdktypes.NewIntegerValue(11),
		}, vs)
	}

	vs, err = db.ListValues(ctx, pids[1])
	if assert.NoError(t, err) {
		assert.Equal(t, map[string]sdktypes.Value{"key": sdktypes.NewIntegerValue(2)}, vs)
	}

	vs, err = db.ListValues(ctx, pids[2])
	if assert.NoError(t, err) {
		assert.Equal(t, map[string]sdktypes.Value{}, vs)
	}

	require.NoError(t, db.SetValue(ctx, pids[0], "key0", sdktypes.NewIntegerValue(100)))

	vs, err = db.ListValues(ctx, pids[0])
	if assert.NoError(t, err) {
		assert.Equal(t, map[string]sdktypes.Value{
			"key0": sdktypes.NewIntegerValue(100),
			"key1": sdktypes.NewIntegerValue(11),
		}, vs)
	}

	v, err := db.GetValue(ctx, pids[0], "key0")
	if assert.NoError(t, err) {
		assert.Equal(t, sdktypes.NewIntegerValue(100), v)
	}

	v, err = db.GetValue(ctx, pids[1], "key")
	if assert.NoError(t, err) {
		assert.Equal(t, sdktypes.NewIntegerValue(2), v)
	}

	v, err = db.GetValue(ctx, pids[2], "key")
	assert.ErrorIs(t, err, sdkerrors.ErrNotFound)
	assert.False(t, v.IsValid())
}
