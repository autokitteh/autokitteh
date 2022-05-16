//go:build unit_norace && !race

package eventsrcsstoregrpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/autokitteh/autokitteh/internal/app/eventsrcsstoregrpcsvc"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsrcsstore"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsrcsstore/eventsrcsstoregorm"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsrcsstore/tests"
)

func newStore(t *testing.T) func() eventsrcsstore.Store {
	return func() eventsrcsstore.Store {
		db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
		require.NoError(t, err)

		as := eventsrcsstoregorm.Store{DB: db}

		require.NoError(t, as.Teardown(context.Background()))
		require.NoError(t, as.Setup(context.Background()))

		svc := eventsrcsstoregrpcsvc.Svc{Store: &as}

		client := eventsrcsstoregrpcsvc.LocalClient{Server: &svc}

		return &Store{Client: &client}
	}
}

func TestAllGRPC(t *testing.T) {
	tests.TestAll(t, newStore(t))
}
