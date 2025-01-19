package dbgorm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func TestDupOrgName(t *testing.T) {
	db, err := New(zaptest.NewLogger(t), &Config{})
	require.NoError(t, err)

	ctx := context.Background()

	require.NoError(t, db.Connect(ctx))
	require.NoError(t, db.Setup(ctx))

	org := sdktypes.NewOrg().WithNewID()
	oid, err := db.CreateOrg(ctx, org)
	require.NoError(t, err)
	require.Equal(t, oid, org.ID())

	_, err = db.CreateOrg(ctx, org)
	require.ErrorIs(t, err, sdkerrors.ErrAlreadyExists)

	dogs := org.WithNewID()
	_, err = db.CreateOrg(ctx, dogs)
	require.NoError(t, err)

	cats := org.WithNewID().WithName(sdktypes.NewSymbol("cats"))
	_, err = db.CreateOrg(ctx, cats)
	require.NoError(t, err)

	_, err = db.CreateOrg(ctx, cats.WithNewID())
	require.ErrorIs(t, err, sdkerrors.ErrAlreadyExists)

	err = db.UpdateOrg(ctx, dogs.WithName(cats.Name()), &sdktypes.FieldMask{Paths: []string{"name"}})
	require.ErrorIs(t, err, sdkerrors.ErrAlreadyExists)

	err = db.UpdateOrg(ctx, dogs.WithName(sdktypes.NewSymbol("dogs")), &sdktypes.FieldMask{Paths: []string{"name"}})
	require.NoError(t, err)
}
