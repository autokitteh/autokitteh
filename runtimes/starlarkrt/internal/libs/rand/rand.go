package rand

import (
	"math/rand"

	"go.starlark.net/starlark"
)

const ModuleName = "rand"

type module struct{ r *rand.Rand }

func LoadModule(seed int64) (starlark.StringDict, error) {
	m := &module{r: rand.New(rand.NewSource(seed))}

	return starlark.StringDict(map[string]starlark.Value{
		"intn": starlark.NewBuiltin("intn", m.intn),
		"seed": starlark.NewBuiltin("seed", m.seed),
	}), nil
}

func (m *module) intn(thread *starlark.Thread, bi *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var n int

	if err := starlark.UnpackArgs(bi.Name(), args, kwargs, "n", &n); err != nil {
		return nil, err
	}

	return starlark.MakeInt(m.r.Intn(n)), nil
}

func (m *module) seed(thread *starlark.Thread, bi *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var seed int64

	if err := starlark.UnpackArgs(bi.Name(), args, kwargs, "seed", &seed); err != nil {
		return nil, err
	}

	m.r.Seed(seed)

	return starlark.None, nil
}
