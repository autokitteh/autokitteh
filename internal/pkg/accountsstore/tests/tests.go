//go:build unit || unit_norace

package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/autokitteh/autokitteh/internal/pkg/accountsstore"
	"go.autokitteh.dev/sdk/api/apiaccount"
)

func TestAll(t *testing.T, as accountsstore.Store) {
	t.Run("create_and_get", func(t *testing.T) { TestCreateAndGet(t, as) })
	t.Run("get_not_found", func(t *testing.T) { TestGetNotFound(t, as) })
	t.Run("update", func(t *testing.T) { TestUpdate(t, as) })
}

func TestCreateAndGet(t *testing.T, as accountsstore.Store) {
	id, err := as.Create(
		context.Background(),
		accountsstore.AutoAccountID,
		(&apiaccount.AccountData{}).
			SetName("test").
			SetMemo(map[string]string{"1": "one"}),
	)
	require.NoError(t, err)

	a, err := as.Get(context.Background(), id)
	require.NoError(t, err)
	require.NotNil(t, a)

	assert.Equal(t, id, a.ID())

	assert.Equal(t, "test", a.Data().Name())
	assert.Equal(t, map[string]string{"1": "one"}, a.Data().Memo())
}

func TestGetNotFound(t *testing.T, as accountsstore.Store) {
	_, err := as.Get(context.Background(), apiaccount.NewAccountID())

	require.True(t, errors.Is(err, accountsstore.ErrNotFound))
}

func TestUpdate(t *testing.T, as accountsstore.Store) {
	id, err := as.Create(
		context.Background(),
		accountsstore.AutoAccountID,

		(&apiaccount.AccountData{}).
			SetName("original").
			SetMemo(map[string]string{"1": "one"}).
			SetEnabled(true),
	)
	require.NoError(t, err)

	a, err := as.Get(context.Background(), id)
	require.NoError(t, err)
	require.NotNil(t, a)

	assert.True(t, a.Data().Enabled())

	require.NoError(t, as.Update(
		context.Background(),
		id,
		(&apiaccount.AccountData{}).SetEnabled(false),
	))

	a, err = as.Get(context.Background(), id)
	require.NoError(t, err)
	require.NotNil(t, a)

	assert.False(t, a.Data().Enabled())
	assert.Equal(t, "original", a.Data().Name())

	require.NoError(t, as.Update(
		context.Background(),
		id,
		(&apiaccount.AccountData{}).SetName("updated").SetEnabled(true),
	))

	a, err = as.Get(context.Background(), id)
	require.NoError(t, err)
	require.NotNil(t, a)

	assert.True(t, a.Data().Enabled())
	assert.Equal(t, "updated", a.Data().Name())
}
