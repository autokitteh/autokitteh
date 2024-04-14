package dbgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetGetSecret(t *testing.T) {
	f := newDBFixture()

	want := map[string]string{"key": "value"}
	err := f.gormdb.SetSecret(f.ctx, "name", want)
	assert.NoError(t, err)

	got, err := f.gormdb.GetSecret(f.ctx, "name")
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestGetSecretError(t *testing.T) {
	f := newDBFixture()

	got, err := f.gormdb.GetSecret(f.ctx, "name")
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestSetAppendGetSecret(t *testing.T) {
	f := newDBFixture()

	data := map[string]string{"key1": "value1"}
	err := f.gormdb.SetSecret(f.ctx, "name", data)
	assert.NoError(t, err)

	err = f.gormdb.AppendSecret(f.ctx, "name", "key2")
	assert.NoError(t, err)

	got, err := f.gormdb.GetSecret(f.ctx, "name")
	assert.NoError(t, err)
	assert.NotEqual(t, "", got["key2"])
}

func TestSetDeleteGetSecret(t *testing.T) {
	f := newDBFixture()

	data := map[string]string{"key": "value"}
	err := f.gormdb.SetSecret(f.ctx, "name", data)
	assert.NoError(t, err)

	err = f.gormdb.DeleteSecret(f.ctx, "name")
	assert.NoError(t, err)

	got, err := f.gormdb.GetSecret(f.ctx, "name")
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestDeleteSecretError(t *testing.T) {
	db := newDBFixture()

	err := db.gormdb.DeleteSecret(db.ctx, "name")
	assert.NoError(t, err)
}
