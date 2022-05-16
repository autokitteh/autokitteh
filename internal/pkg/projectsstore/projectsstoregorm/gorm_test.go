//go:build unit_norace && !race

package projectsstoregorm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"gitlab.com/softkitteh/autokitteh/internal/pkg/accountsstore/accountsstoregorm"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/projectsstore/tests"
)

func TestAllSQLite(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)

	as := accountsstoregorm.Store{DB: db}
	ps := Store{DB: db, AccountsStore: &as}

	require.NoError(t, as.Teardown(context.Background()))
	require.NoError(t, ps.Teardown(context.Background()))
	require.NoError(t, as.Setup(context.Background()))
	require.NoError(t, ps.Setup(context.Background()))

	tests.TestAll(t, &ps, &as)
}
