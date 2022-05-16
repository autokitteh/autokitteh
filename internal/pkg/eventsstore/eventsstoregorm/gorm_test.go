//go:build unit_norace && !race

package eventsstoregorm

import (
	"context"
	"flag"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/autokitteh/autokitteh/internal/pkg/eventsstore"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsstore/tests"
)

var dburl = flag.String("dburl", "file::memory:", "sqlite db url")

func newStore(t *testing.T) func() eventsstore.Store {
	return func() eventsstore.Store {
		db, err := gorm.Open(sqlite.Open(*dburl), &gorm.Config{})
		require.NoError(t, err)

		es := Store{DB: db}

		require.NoError(t, es.Teardown(context.Background()))
		require.NoError(t, es.Setup(context.Background()))

		return &es
	}
}

func TestAll(t *testing.T) {
	tests.TestAll(t, newStore(t))
}
