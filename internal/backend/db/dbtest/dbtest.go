package dbtest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbfactory"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func NewTestDB(t *testing.T, objs ...sdktypes.Object) db.DB {
	tdb, err := dbfactory.New(zaptest.NewLogger(t), dbfactory.Configs.Test)
	require.NoError(t, err)

	ctx := context.Background()

	require.NoError(t, tdb.Connect(ctx))
	require.NoError(t, tdb.Setup(ctx))

	require.NoError(t, db.Populate(ctx, tdb, objs...))

	return tdb
}
