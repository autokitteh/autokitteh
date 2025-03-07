package sdktypes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type myStruct struct {
	A string
	B string `var:"b"`
	C string `var:"c,secret"`
	D string `var:",secret"`
	E string `var:"secret"` // Legacy behavior: no new name, just secret.
}

func TestEncodeVars(t *testing.T) {
	s := myStruct{"1", "2", "3", "4", "5"}
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

func TestDecode(t *testing.T) {
	vs := NewVars(
		NewVar(NewSymbol("A")).SetValue("1"),
		NewVar(NewSymbol("b")).SetValue("2"),
		NewVar(NewSymbol("c")).SetValue("3").SetSecret(true),
		NewVar(NewSymbol("D")).SetValue("4").SetSecret(true),
		NewVar(NewSymbol("E")).SetValue("5").SetSecret(true),
		NewVar(NewSymbol("irrelevant")).SetValue("6"),
	)

	s := new(myStruct)
	vs.Decode(s)

	assert.Equal(t, "1", s.A)
	assert.Equal(t, "2", s.B)
	assert.Equal(t, "3", s.C)
	assert.Equal(t, "4", s.D)
	assert.Equal(t, "5", s.E)
}
