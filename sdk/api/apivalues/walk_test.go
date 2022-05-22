//go:build unit

package apivalues

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMutableWalk(t *testing.T) {
	v := Struct(
		Symbol("ctor"),
		map[string]*Value{
			"s": String("woof"),
			"f": MustNewValue(CallValue{Name: "test", ID: "C0001"}),
			"l": List(Integer(1)),
		},
	)

	if assert.NoError(
		t,
		Walk(
			v,
			func(curr, _ *Value, _ Role) error {
				if _, ok := curr.Get().(IntegerValue); ok {
					next, _ := Inc(curr, 1)
					*curr = *next
				} else if _, ok := curr.Get().(StringValue); ok {
					*curr = *String("meow")
				} else {
					_ = SetCallIssuer(curr, "issuer")
				}

				return nil
			},
		),
	) {
		assert.Equal(t, "issuer", v.Get().(StructValue).Fields["f"].Get().(CallValue).Issuer)
		assert.Equal(t, "meow", v.Get().(StructValue).Fields["s"].Get().(StringValue).String())
		assert.Equal(t, List(Integer(2)), v.Get().(StructValue).Fields["l"])
	}
}
