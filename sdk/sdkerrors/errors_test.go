package sdkerrors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var one = 1

func TestIgnoreNotFoundErr(t *testing.T) {
	x, err := IgnoreNotFoundErr[int](&one, ErrNotFound)
	if assert.NoError(t, err) {
		assert.Equal(t, x, &one)
	}

	x, err = IgnoreNotFoundErr[int](&one, ErrConflict)
	if assert.Equal(t, err, ErrConflict) {
		assert.Nil(t, x)
	}
}
