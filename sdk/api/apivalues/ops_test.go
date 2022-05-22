//go:build unit

package apivalues

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetIssuer(t *testing.T) {
	cv := MustNewValue(CallValue{Name: "test", ID: "C0001"})

	if assert.NoError(t, SetCallIssuer(cv, "issuer")) {
		assert.Equal(t, "issuer", cv.Get().(CallValue).Issuer)
	}
}
