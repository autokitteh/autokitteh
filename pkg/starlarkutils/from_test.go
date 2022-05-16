package starlarkutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func TestFromStarlark(t *testing.T) {
	intptr := func(i int64) *int64 { return &i }
	strptr := func(s string) *string { return &s }

	type subStruct struct {
		S string
	}

	type testStruct struct {
		X, Y int
		L    []int
		S    subStruct
		P    *subStruct
	}

	outStruct := testStruct{X: 1, Y: 2, L: []int{1, 2}, S: subStruct{S: "meow"}, P: &subStruct{S: "woof"}}

	tests := []struct {
		name     string
		expected interface{}
		out      interface{}
		in       starlark.Value
	}{
		{
			name:     "int",
			expected: intptr(42),
			out:      intptr(0),
			in:       starlark.MakeInt(42),
		},
		{
			name:     "str",
			expected: strptr("meow"),
			out:      strptr(""),
			in:       starlark.String("meow"),
		},
		{
			name:     "none list",
			expected: &([]int{1, 2, 3}),
			out:      &([]int{1, 2, 3}),
			in:       starlark.None,
		},
		{
			name:     "list",
			expected: &([]int{1, 2, 3}),
			out:      &([]int{}),
			in:       starlark.NewList([]starlark.Value{starlark.MakeInt(1), starlark.MakeInt(2), starlark.MakeInt(3)}),
		},
		{
			name:     "none array",
			expected: &([3]int{1, 2, 3}),
			out:      &([3]int{1, 2, 3}),
			in:       starlark.None,
		},
		{
			name:     "array",
			expected: &([3]int{1, 2, 3}),
			out:      &([3]int{}),
			in:       starlark.NewList([]starlark.Value{starlark.MakeInt(1), starlark.MakeInt(2), starlark.MakeInt(3)}),
		},
		{
			name:     "none map",
			expected: &(map[string]int{"one": 1, "two": 2}),
			out:      &(map[string]int{"one": 1, "two": 2}),
			in:       starlark.None,
		},
		{
			name:     "map",
			expected: &(map[string]int{"one": 1, "two": 2}),
			out:      &(map[string]int{}),
			in: func() starlark.Value {
				testDict := starlark.NewDict(2)
				_ = testDict.SetKey(starlark.String("one"), starlark.MakeInt(1))
				_ = testDict.SetKey(starlark.String("two"), starlark.MakeInt(2))
				return testDict
			}(),
		},
		{
			name: "none struct",
			expected: func() interface{} {
				o := outStruct
				return &o
			}(),
			out: &outStruct,
			in:  starlark.None,
		},
		{
			name:     "struct",
			expected: &outStruct,
			out:      &testStruct{},
			in: starlarkstruct.FromStringDict(
				starlark.None,
				map[string]starlark.Value{
					"x": starlark.MakeInt(1),
					"y": starlark.MakeInt(2),
					"l": starlark.NewList([]starlark.Value{starlark.MakeInt(1), starlark.MakeInt(2)}),
					"s": starlarkstruct.FromStringDict(
						starlark.None,
						map[string]starlark.Value{
							"s": starlark.String("meow"),
						},
					),
					"p": starlarkstruct.FromStringDict(
						starlark.None,
						map[string]starlark.Value{
							"s": starlark.String("woof"),
						},
					),
				},
			),
		},
		{
			name:     "struct from map",
			expected: &outStruct,
			out:      &testStruct{},
			in: func() starlark.Value {
				d := starlark.NewDict(3)
				_ = d.SetKey(starlark.String("x"), starlark.MakeInt(1))
				_ = d.SetKey(starlark.String("y"), starlark.MakeInt(2))
				_ = d.SetKey(starlark.String("l"), starlark.NewList([]starlark.Value{starlark.MakeInt(1), starlark.MakeInt(2)}))

				dd := starlark.NewDict(1)
				_ = dd.SetKey(starlark.String("s"), starlark.String("meow"))
				_ = d.SetKey(starlark.String("s"), dd)

				dd = starlark.NewDict(1)
				_ = dd.SetKey(starlark.String("s"), starlark.String("woof"))
				_ = d.SetKey(starlark.String("p"), dd)
				return d
			}(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := FromStarlark(test.in, test.out)
			if !assert.NoError(t, err) {
				return
			}

			assert.Equal(t, test.expected, test.out)
		})
	}
}
