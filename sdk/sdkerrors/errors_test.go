package sdkerrors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIgnoreNotFoundErr(t *testing.T) {
	x, err := IgnoreNotFoundErr[int](1, ErrNotFound)
	if assert.NoError(t, err) {
		assert.Equal(t, x, 1)
	}

	x, err = IgnoreNotFoundErr[int](1, ErrConflict)
	if assert.Equal(t, err, ErrConflict) {
		assert.Zero(t, x)
	}
}
