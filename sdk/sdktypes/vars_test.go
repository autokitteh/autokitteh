package sdktypes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type MyStruct struct {
	A string
	B string `var:"b"`
	C string `var:"c,secret"`
	D string `var:",secret"`
	E string `var:"secret"` // Legacy behavior: no new name, just secret.
}

func TestEncodeVars(t *testing.T) {
	s := MyStruct{"1", "2", "3", "4", "5"}
	vs := EncodeVars(s)

	assert.False(t, vs.Has(NewSymbol("")))

	assert.True(t, vs.Has(NewSymbol("A")))
	assert.False(t, vs.GetByString("A").IsSecret())

	assert.False(t, vs.Has(NewSymbol("B")))
	assert.True(t, vs.Has(NewSymbol("b")))
	assert.False(t, vs.GetByString("b").IsSecret())

	assert.False(t, vs.Has(NewSymbol("C")))
	assert.True(t, vs.Has(NewSymbol("c")))
	assert.True(t, vs.GetByString("c").IsSecret())

	assert.True(t, vs.Has(NewSymbol("D")))
	assert.True(t, vs.GetByString("D").IsSecret())

	// Legacy behavior: no new name, just secret.
	assert.False(t, vs.Has(NewSymbol("secret")))
	assert.True(t, vs.Has(NewSymbol("E")))
	assert.True(t, vs.GetByString("E").IsSecret())
}
