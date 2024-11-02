package sdkerrors

import (
	"errors"
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

func TestErrorType(t *testing.T) {
	tests := []struct {
		err error
		typ string
	}{
		{ErrNotFound, "not_found"},
		{NewRetryableError(ErrNotFound), "retryable_not_found"},
		{NewInvalidArgumentError("meow"), "invalid_argument"},
		{errors.New("woof"), "unknown"},
	}

	for _, test := range tests {
		t.Run(test.typ, func(t *testing.T) {
			assert.Equal(t, ErrorType(test.err), test.typ)
		})
	}
}
