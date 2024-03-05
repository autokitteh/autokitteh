package sdktypes_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func TestExecutorID(t *testing.T) {
	var xid sdktypes.ExecutorID
	assert.False(t, xid.IsValid())
	assert.Equal(t, "", xid.String())
	assert.Equal(t, "", xid.Kind())

	rid := sdktypes.NewRunID()
	iid := sdktypes.NewIntegrationID()

	rxid := sdktypes.NewExecutorID(rid)
	assert.True(t, rxid.IsRunID())
	assert.False(t, rxid.IsIntegrationID())
	assert.Equal(t, rid, rxid.ToRunID())
	assert.Equal(t, rid.String(), rxid.String())
	assert.NotEqual(t, iid.String(), rxid.String())
	assert.False(t, rxid.ToIntegrationID().IsValid())

	ixid := sdktypes.NewExecutorID(iid)
	assert.True(t, ixid.IsIntegrationID())
	assert.False(t, ixid.IsRunID())
	assert.Equal(t, iid, ixid.ToIntegrationID())
	assert.Equal(t, iid.String(), ixid.String())
	assert.NotEqual(t, rid.String(), ixid.String())
	assert.False(t, ixid.ToRunID().IsValid())

	assert.NotEqual(t, rxid, ixid)

	xid = rxid
	assert.True(t, xid.IsValid())
	assert.Equal(t, rxid, xid)
	assert.NotEqual(t, ixid, xid)
}
