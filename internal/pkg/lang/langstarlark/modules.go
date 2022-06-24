package langstarlark

import (
	"fmt"
	"time"

	"github.com/qri-io/starlib"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"

	"github.com/autokitteh/autokitteh/internal/pkg/lang/langstarlark/modules/parsecmd"
	"github.com/autokitteh/autokitteh/internal/pkg/lang/langstarlark/modules/reflect"
	"github.com/autokitteh/autokitteh/internal/pkg/lang/langstarlark/modules/txtar"
	"github.com/autokitteh/autokitteh/internal/pkg/lang/langstarlark/starlarktest"
	"github.com/autokitteh/starlarkutils"
)

var (
	DeterministicBuiltinModules = map[string]func() (starlark.StringDict, error){
		"reflect":  func() (starlark.StringDict, error) { return reflect.Load(), nil },
		"parsecmd": func() (starlark.StringDict, error) { return parsecmd.Load(), nil },
		"txtar":    func() (starlark.StringDict, error) { return txtar.Load(), nil },
	}
	PermissiveBuiltinModules = make(map[string]func() (starlark.StringDict, error))

	DeterministicBuiltinValues = map[string]starlark.Value{
		"struct": starlark.NewBuiltin("struct", starlarkstruct.Make),
		"module": starlark.NewBuiltin("module", starlarkstruct.MakeModule),
		"symbol": starlark.NewBuiltin("gensym", starlarkutils.GenSymbol),
		"fail":   starlarktest.FailBuiltin,
		"assert": starlarktest.AssertBuiltin,
		"catch":  starlarktest.CatchBuiltin,
	}
	PermissiveBuiltinValues map[string]starlark.Value
)

func init() {
	PermissiveBuiltinValues = make(map[string]starlark.Value, len(DeterministicBuiltinValues)+1)
	for k, v := range DeterministicBuiltinModules {
		PermissiveBuiltinModules[k] = v
	}

	// TODO: This should really belong elsewhere. Why isn't this in starlark's time?
	PermissiveBuiltinValues["sleep"] = starlark.NewBuiltin(
		"sleep",
		func(th *starlark.Thread, bi *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			ctx := getTLSContext(th)

			var t number
			if err := starlark.UnpackArgs(bi.Name(), args, kwargs, "t", &t); err != nil {
				return nil, err
			}

			select {
			case <-time.After(time.Duration(float64(t.AsFloat()) * float64(time.Second))):
				return starlark.None, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		},
	)

	dmods := []string{
		"gzip",
		"xslx",
		"html",
		"zipfile",
		"re",
		"encoding/base64",
		"encoding/csv",
		"encoding/json",
		"encoding/yaml",
		"geo",
		"math",
		"hash",
		"dataframe",
	}

	pmods := []string{
		"time",
		"http",
	}

	load := func(names []string, dst map[string]func() (starlark.StringDict, error)) {
		for _, name := range names {
			func(name string) {
				dst[name] = func() (starlark.StringDict, error) {
					return starlib.Loader(nil, fmt.Sprintf(name+".star"))
				}
			}(name)
		}
	}

	load(dmods, DeterministicBuiltinModules)

	load(dmods, PermissiveBuiltinModules)
	load(pmods, PermissiveBuiltinModules)
}
