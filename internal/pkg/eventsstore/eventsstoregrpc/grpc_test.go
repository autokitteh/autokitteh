//go:build unit_norace && !race

package eventsstoregrpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/autokitteh/autokitteh/internal/app/eventsstoregrpcsvc"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsstore"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsstore/eventsstoregorm"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsstore/tests"
)

func newStore(t *testing.T) func() eventsstore.Store {
	return func() eventsstore.Store {
		db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
		require.NoError(t, err)

		as := eventsstoregorm.Store{DB: db}

		require.NoError(t, as.Teardown(context.Background()))
		require.NoError(t, as.Setup(context.Background()))

		svc := eventsstoregrpcsvc.Svc{Store: &as}

		client := eventsstoregrpcsvc.LocalClient{Server: &svc}

		return &Store{Client: &client}
	}
}

func TestAllGRPC(t *testing.T) {
	tests.TestAll(t, newStore(t))
}
