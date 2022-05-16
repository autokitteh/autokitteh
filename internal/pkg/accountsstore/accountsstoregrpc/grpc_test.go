//go:build unit_norace && !race

package accountsstoregrpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"gitlab.com/softkitteh/autokitteh/internal/app/accountsstoregrpcsvc"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/accountsstore/accountsstoregorm"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/accountsstore/tests"
)

func TestAllGRPC(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)

	as := accountsstoregorm.Store{DB: db}

	require.NoError(t, as.Teardown(context.Background()))
	require.NoError(t, as.Setup(context.Background()))

	svc := accountsstoregrpcsvc.Svc{Store: &as}

	client := accountsstoregrpcsvc.LocalClient{Server: &svc}

	grpcas := Store{Client: &client}

	tests.TestAll(t, &grpcas)
}
