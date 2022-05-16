//go:build unit || unit_norace

package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/softkitteh/autokitteh/internal/pkg/accountsstore"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiaccount"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiprogram"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiproject"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/projectsstore"
)

var path = apiprogram.MustParsePathString("fs:/tmp")

func TestAll(t *testing.T, ps projectsstore.Store, as accountsstore.Store) {
	t.Run("create_and_get", func(t *testing.T) { TestCreateAndGet(t, ps, as) })
	t.Run("get_not_found", func(t *testing.T) { TestGetNotFound(t, ps) })
	t.Run("update", func(t *testing.T) { TestUpdate(t, ps, as) })
	t.Run("has_account", func(t *testing.T) { TestHasAccount(t, ps, as) })
}

func TestHasAccount(t *testing.T, ps projectsstore.Store, as accountsstore.Store) {
	_, err := ps.Create(
		context.Background(),
		apiaccount.NewAccountID(),
		projectsstore.AutoProjectID,
		&apiproject.ProjectData{},
	)
	require.Error(t, err)
	require.True(t, errors.Is(err, projectsstore.ErrInvalidAccount))
}

func TestCreateAndGet(t *testing.T, ps projectsstore.Store, as accountsstore.Store) {
	aid, err := as.Create(context.Background(), accountsstore.AutoAccountID, nil)
	require.NoError(t, err)

	id, err := ps.Create(
		context.Background(),
		aid,
		projectsstore.AutoProjectID,
		(&apiproject.ProjectData{}).
			SetName("test").
			SetMemo(map[string]string{"1": "one"}).
			SetMainPath(path),
	)
	require.NoError(t, err)

	a, err := ps.Get(context.Background(), id)
	require.NoError(t, err)
	require.NotNil(t, a)

	assert.Equal(t, id, a.ID())
	assert.Equal(t, aid, a.AccountID())

	d := a.Data()
	assert.Equal(t, "test", d.Name())
	assert.Equal(t, map[string]string{"1": "one"}, d.Memo())
}

func TestGetNotFound(t *testing.T, ps projectsstore.Store) {
	_, err := ps.Get(context.Background(), apiproject.NewProjectID())

	require.True(t, errors.Is(err, projectsstore.ErrNotFound))
}

func TestUpdate(t *testing.T, ps projectsstore.Store, as accountsstore.Store) {
	aid, err := as.Create(context.Background(), accountsstore.AutoAccountID, nil)
	require.NoError(t, err)

	id, err := ps.Create(
		context.Background(),
		aid,
		projectsstore.AutoProjectID,
		(&apiproject.ProjectData{}).
			SetName("original").
			SetMemo(map[string]string{"1": "one"}).
			SetEnabled(true).
			SetMainPath(path),
	)
	require.NoError(t, err)

	a, err := ps.Get(context.Background(), id)
	require.NoError(t, err)
	require.NotNil(t, a)

	assert.True(t, a.Data().Enabled())

	require.NoError(t, ps.Update(
		context.Background(),
		id,
		&apiproject.ProjectData{},
	))

	a, err = ps.Get(context.Background(), id)
	require.NoError(t, err)
	require.NotNil(t, a)

	assert.False(t, a.Data().Enabled())
	assert.Equal(t, "original", a.Data().Name())

	require.NoError(t, ps.Update(
		context.Background(),
		id,
		(&apiproject.ProjectData{}).SetName("updated").SetEnabled(true),
	))

	a, err = ps.Get(context.Background(), id)
	require.NoError(t, err)
	require.NotNil(t, a)

	assert.True(t, a.Data().Enabled())
	assert.Equal(t, "updated", a.Data().Name())
}
