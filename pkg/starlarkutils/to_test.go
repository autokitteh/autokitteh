package starlarkutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"go.starlark.net/syntax"
)

func TestToStarlark(t *testing.T) {
	tests := []struct {
		name string
		in   interface{}
		out  starlark.Value
	}{
		{
			name: "nil",
			out:  starlark.None,
		},
		{
			name: "42",
			in:   42,
			out:  starlark.MakeInt(42),
		},
		{
			name: "meow",
			in:   "meow",
			out:  starlark.String("meow"),
		},
		{
			name: "true",
			in:   true,
			out:  starlark.Bool(true),
		},
		{
			name: "1.2",
			in:   1.2,
			out:  starlark.Float(1.2),
		},
		{
			name: "slice 1,2,3",
			in:   []int{1, 2, 3},
			out:  starlark.NewList([]starlark.Value{starlark.MakeInt(1), starlark.MakeInt(2), starlark.MakeInt(3)}),
		},
		{
			name: "array 1,2,3",
			in:   [3]int{1, 2, 3},
			out:  starlark.NewList([]starlark.Value{starlark.MakeInt(1), starlark.MakeInt(2), starlark.MakeInt(3)}),
		},
		{
			name: "map",
			in:   map[string]interface{}{"one": 1, "l": []string{"meow", "woof"}},
			out: func() starlark.Value {
				dict := starlark.NewDict(2)
				_ = dict.SetKey(starlark.String("one"), starlark.MakeInt(1))
				_ = dict.SetKey(starlark.String("l"), starlark.NewList([]starlark.Value{starlark.String("meow"), starlark.String("woof")}))

				return dict
			}(),
		},
		{
			name: "struct",
			in: func() interface{} {
				type embedded struct{ E int }
				type testStruct struct {
					embedded
					X, Y, unexported int
				}

				return testStruct{X: 1, Y: 2, unexported: 9}
			}(),
			out: starlarkstruct.FromStringDict(
				starlark.String("test_struct"),
				map[string]starlark.Value{
					"e": starlark.MakeInt(0),
					"x": starlark.MakeInt(1),
					"y": starlark.MakeInt(2),
				},
			),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out, err := ToStarlark(test.in)
			if !assert.NoError(t, err) {
				return
			}

			eq, err := starlark.Compare(syntax.EQL, test.out, out)
			if assert.NoError(t, err) {
				if !assert.True(t, eq) {
					t.Log(test.out)
					t.Log(out)
				}
			}
		})
	}
}
