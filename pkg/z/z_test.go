//go:build unit

package z

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNop(t *testing.T) {
	assert.NotNil(t, Z(nil))
}

func TestReal(t *testing.T) {
	l, _ := zap.NewDevelopment()
	z := l.Sugar()
	assert.Equal(t, z, Z(z))
}
