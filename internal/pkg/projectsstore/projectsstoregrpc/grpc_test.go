//go:build unit_norace && !race

package projectsstoregrpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/autokitteh/autokitteh/internal/app/projectsstoregrpcsvc"
	"github.com/autokitteh/autokitteh/internal/pkg/accountsstore/accountsstoregorm"
	"github.com/autokitteh/autokitteh/internal/pkg/projectsstore/projectsstoregorm"
	"github.com/autokitteh/autokitteh/internal/pkg/projectsstore/tests"
)

func TestAllGRPC(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)

	as := accountsstoregorm.Store{DB: db}
	ps := projectsstoregorm.Store{DB: db, AccountsStore: &as}

	require.NoError(t, ps.Setup(context.Background()))
	require.NoError(t, as.Setup(context.Background()))

	svc := projectsstoregrpcsvc.Svc{Store: &ps}

	client := projectsstoregrpcsvc.LocalClient{Server: &svc}

	grpcps := Store{Client: &client}

	tests.TestAll(t, &grpcps, &as)
}
