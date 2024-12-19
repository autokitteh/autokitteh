package opapolicytest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authz"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm"
	"go.autokitteh.dev/autokitteh/internal/backend/policy/opapolicy"
)

func InitAuthzTest(t *testing.T, configPath string) (db.DB, context.Context) {
	ctx := context.Background()

	l := zaptest.NewLogger(t)
	db, err := dbgorm.New(l, nil)
	require.NoError(t, err)

	require.NoError(t, db.Connect(ctx))
	require.NoError(t, db.Setup(ctx))

	decide, err := opapolicy.New(&opapolicy.Config{ConfigPath: configPath}, l, db)
	require.NoError(t, err)

	check := authz.NewPolicyCheckFunc(zaptest.NewLogger(t), db, decide)

	return db, authz.ContextWithCheckFunc(ctx, check)
}
