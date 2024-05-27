package dbgorm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompressJSON(t *testing.T) {
	db := &gormdb{
		compressionThreshold: 1024,
	}

	u, c, err := db.compressJSON("meow")
	if assert.NoError(t, err) {
		assert.Equal(t, string(u), `"meow"`)
		assert.Nil(t, c)
	}

	bs := []byte(strings.Repeat("meow", 512))

	u, c, err = db.compressJSON(bs)
	if assert.NoError(t, err) {
		assert.Nil(t, u)
		assert.NotNil(t, c)
		t.Logf("compressed: %d -> %d", len(bs), len(c))
	}
}
