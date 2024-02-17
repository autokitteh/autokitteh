package sdktypes

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUUIDGenerator(t *testing.T) {
	uuid := UUIDGenerator()
	assert.Len(t, uuid, 32)
	_, err := hex.DecodeString(uuid)
	assert.NoError(t, err)
}

func TestSequentialIDGeneratorForTesting(t *testing.T) {
	g := NewSequentialIDGeneratorForTesting(42)

	assert.NotNil(t, g)

	for i := uint8(42); i < 42+10; i++ {
		s := g()

		assert.Len(t, s, 32)
		bs, err := hex.DecodeString(s)
		assert.NoError(t, err)

		assert.Equal(t, bs[15], i+1)
	}
}
