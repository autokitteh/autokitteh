package sdkmodule

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func TestSimpleUnpack(t *testing.T) {
	assert.NoError(t, UnpackArgs(nil, nil))

	var (
		i  int
		st struct {
			X int
		}
	)

	assert.Error(t, UnpackArgs(nil, nil, "i", &i, &st))

	if assert.NoError(t, UnpackArgs([]sdktypes.Value{sdktypes.NewIntegerValue(42)}, nil, "i", &i)) {
		assert.Equal(t, 42, i)
		assert.Zero(t, st)
	}

	var s string

	if assert.NoError(t, UnpackArgs([]sdktypes.Value{sdktypes.NewIntegerValue(64)}, map[string]sdktypes.Value{
		"s": sdktypes.NewStringValue("meow"),
	}, "i", &i, "s=", &s)) {
		assert.Equal(t, 64, i)
		assert.Equal(t, "meow", s)
		assert.Zero(t, st)
	}
}

func TestUnpackFlat(t *testing.T) {
	var (
		i  int
		st struct {
			X int     `json:"x"`
			Y *string `json:"y"`
			S struct {
				Z int `json:"z"`
			} `json:"s"`
			Sptr *struct {
				Z int `json:"z"`
			} `json:"sptr"`
		}
	)

	assert.NoError(t, UnpackArgs([]sdktypes.Value{
		sdktypes.NewIntegerValue(42),
	}, map[string]sdktypes.Value{
		"x": sdktypes.NewIntegerValue(64),
		"y": sdktypes.NewStringValue("meow"),
		"s": sdktypes.NewDictValueFromStringMap(map[string]sdktypes.Value{
			"z": sdktypes.NewIntegerValue(128),
		}),
	}, "i", &i, &st))

	assert.Equal(t, 42, i)
	assert.Equal(t, 64, st.X)
	assert.Equal(t, "meow", *st.Y)
	assert.Equal(t, 128, st.S.Z)
	assert.Zero(t, st.Sptr)
}

func TestUnpack(t *testing.T) {
	one := sdktypes.NewIntegerValue(1)
	two := sdktypes.NewIntegerValue(2)
	three := sdktypes.NewIntegerValue(3)
	meow := sdktypes.NewStringValue("meow")

	type outs struct {
		I  int `json:"j"`
		S  string
		Xs []int          `json:",omitempty"`
		M  map[string]int `json:"m,omitempty"`
	}

	var dsts outs

	tests := []struct {
		name   string
		args   []sdktypes.Value
		kwargs map[string]sdktypes.Value
		dsts   []any
		want   outs
	}{
		{
			name: "empty",
		},
		{
			name: "positionals",
			args: []sdktypes.Value{
				one,
				meow,
			},
			dsts: []any{"i", &dsts.I, "s", &dsts.S},
			want: outs{
				I: 1,
				S: "meow",
			},
		},
		{
			name: "kwargs",
			kwargs: map[string]sdktypes.Value{
				"i": one,
				"s": meow,
			},
			dsts: []any{"i", &dsts.I, "s=", &dsts.S},
			want: outs{
				I: 1,
				S: "meow",
			},
		},
		{
			name: "mixed",
			args: []sdktypes.Value{one},
			kwargs: map[string]sdktypes.Value{
				"s": meow,
			},
			dsts: []any{"i", &dsts.I, "s=", &dsts.S},
			want: outs{
				I: 1,
				S: "meow",
			},
		},
		{
			name: "varargs",
			args: []sdktypes.Value{one, two, three},
			dsts: []any{"i", &dsts.I, "*args", &dsts.Xs},
			want: outs{
				I:  1,
				Xs: []int{2, 3},
			},
		},
		{
			name: "varargs-mixed",
			args: []sdktypes.Value{two, three},
			kwargs: map[string]sdktypes.Value{
				"one": one,
				"two": two,
			},
			dsts: []any{"*args", &dsts.Xs, "**kwargs", &dsts.M},
			want: outs{
				Xs: []int{2, 3},
				M:  map[string]int{"one": 1, "two": 2},
			},
		},
		{
			name: "struct",
			kwargs: map[string]sdktypes.Value{
				"j": one,
				"S": meow,
				"m": sdktypes.NewDictValueFromStringMap(map[string]sdktypes.Value{"one": one, "two": two}),
			},
			dsts: []any{&dsts},
			want: outs{
				S: "meow",
				I: 1,
				M: map[string]int{
					"one": 1,
					"two": 2,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dsts = outs{}
			assert.NoError(t, UnpackArgs(test.args, test.kwargs, test.dsts...))
			assert.Equal(t, test.want, dsts)
		})
	}
}
