package kittehs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorWithPrefix(t *testing.T) {
	assert.NoError(t, ErrorWithPrefix("nil error", nil))

	err := ErrorWithPrefix("non-nil error", assert.AnError)
	assert.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
	assert.ErrorContains(t, err, "non-nil error")
}

func TestErrorWithValue(t *testing.T) {
	assert.NoError(t, ErrorWithValue("nil error", nil))

	err := ErrorWithValue("non-nil error", assert.AnError)
	assert.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
	assert.ErrorContains(t, err, "non-nil error")
}
