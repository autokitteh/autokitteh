package sdktypes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNonzeroMessage(t *testing.T) {
	assert.Error(t, nonzeroMessage(&CallFramePB{}))
	assert.NoError(t, nonzeroMessage(&CallFramePB{Name: "test"}))
}

func TestErrorWithValue(t *testing.T) {
	assert.NoError(t, errorForValue("nil error", nil))

	err := errorForValue("non-nil error", assert.AnError)
	assert.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
	assert.ErrorContains(t, err, "non-nil error")
}
