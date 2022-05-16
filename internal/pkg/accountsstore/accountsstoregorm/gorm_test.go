//go:build unit_norace && !race

package accountsstoregorm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"gitlab.com/softkitteh/autokitteh/internal/pkg/accountsstore/tests"
)

func TestAllSQLite(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)

	as := Store{DB: db}

	require.NoError(t, as.Teardown(context.Background()))
	require.NoError(t, as.Setup(context.Background()))

	tests.TestAll(t, &as)
}
