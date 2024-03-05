package sdktypes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNonzeroMessage(t *testing.T) {
	assert.Error(t, nonzeroMessage(&CallFramePB{}))
	assert.NoError(t, nonzeroMessage(&CallFramePB{Name: "test"}))
}
