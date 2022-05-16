//go:build unit_norace && !race

package eventsrcsstoregorm

import (
	"context"
	"flag"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/autokitteh/autokitteh/internal/pkg/eventsrcsstore"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsrcsstore/tests"
)

var dburl = flag.String("dburl", "file::memory:", "sqlite db url")

func newStore(t *testing.T) func() eventsrcsstore.Store {
	return func() eventsrcsstore.Store {
		db, err := gorm.Open(sqlite.Open(*dburl), &gorm.Config{})
		require.NoError(t, err)

		es := Store{DB: db}

		require.NoError(t, es.Teardown(context.Background()))
		require.NoError(t, es.Setup(context.Background()))

		return &es
	}
}

func TestAllGORM(t *testing.T) {
	tests.TestAll(t, newStore(t))
}
