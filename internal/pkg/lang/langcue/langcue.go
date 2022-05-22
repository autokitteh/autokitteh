package langcue

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/autokitteh/autokitteh/internal/pkg/lang"
	"github.com/autokitteh/autokitteh/internal/pkg/lang/langtools"
	"github.com/autokitteh/autokitteh/sdk/api/apilang"
	"github.com/autokitteh/autokitteh/sdk/api/apiprogram"
	"github.com/autokitteh/autokitteh/sdk/api/apivalues"
	L "github.com/autokitteh/autokitteh/pkg/l"
)

type decoderFunc func([]byte, interface{}) error

type langvalues struct {
	name     string
	dataOnly bool
	decoder  decoderFunc
}

var compilerVersion = "0"

func init() {
	Register(langtools.PermissiveCatalog)
	Register(langtools.DeterministicCatalog)
}

var (
	NewJSONDataLang = factory(true, UnmarshalJSON)
	NewJSONProgLang = factory(false, UnmarshalJSON)
	NewYAMLDataLang = factory(true, UnmarshalYAML)
	NewYAMLProgLang = factory(false, UnmarshalYAML)
	NewCueDataLang  = factory(true, UnmarshalCue)
	NewCueProgLang  = factory(false, UnmarshalCue)
)

func Register(cat lang.Catalog) {
	cat.Register("json-program", lang.CatalogLang{
		New:  NewJSONProgLang,
		Exts: []string{"kitteh.json"},
	})

	cat.Register("json-data", lang.CatalogLang{
		New:  NewJSONDataLang,
		Exts: []string{"json"},
	})

	cat.Register("yaml-program", lang.CatalogLang{
		New:  NewYAMLProgLang,
		Exts: []string{"kitteh.yaml", "kitteh.yml"},
	})

	cat.Register("yaml-data", lang.CatalogLang{
		New:  NewYAMLDataLang,
		Exts: []string{"yaml", "yml"},
	})

	cat.Register("cue-program", lang.CatalogLang{
		New:  NewCueProgLang,
		Exts: []string{"kitteh.cue"},
	})

	cat.Register("cue-data", lang.CatalogLang{
		New:  NewCueDataLang,
		Exts: []string{"cue"},
	})
}

func factory(dataOnly bool, decoder decoderFunc) func(_ L.L, name string) (lang.Lang, error) {
	return func(_ L.L, name string) (lang.Lang, error) {
		return &langvalues{name: name, dataOnly: dataOnly, decoder: decoder}, nil
	}
}

func (*langvalues) IsCompilerVersionSupported(_ context.Context, v string) (bool, error) {
	return v == compilerVersion, nil
}

func (*langvalues) GetModuleDependencies(_ context.Context, mod *apiprogram.Module) ([]*apiprogram.Path, error) {
	var c compiled

	if err := c.decode(mod.CompiledCode()); err != nil {
		return nil, err
	}

	paths := make([]*apiprogram.Path, 0, len(c.Modules))
	visited := make(map[string]bool, len(c.Modules))

	for _, s := range c.Modules {
		if visited[s.Path.String()] {
			continue
		}

		visited[s.Path.String()] = true

		paths = append(paths, s.Path)
	}

	return paths, nil
}

func (l *langvalues) CompileModule(
	_ context.Context,
	path *apiprogram.Path,
	src []byte,
	_ []string,
) (*apiprogram.Module, error) {
	var c *compiled

	if l.dataOnly {
		var m map[string]interface{}

		if err := l.decoder(src, &m); err != nil {
			return nil, fmt.Errorf("unmarshal: %w", err)
		}

		dst, err := apivalues.Wrap(m)
		if err != nil {
			return nil, fmt.Errorf("unwrap: %w", err)
		}

		dict, ok := dst.Get().(apivalues.DictValue)
		if !ok {
			return nil, fmt.Errorf("data must have a dictionary as root")
		}

		c = &compiled{
			Consts: make(map[string]*apivalues.Value),
		}

		dict.ToStringValuesMap(c.Consts)
	} else {
		var (
			prog prog
			err  error
		)

		if err = l.decoder(src, &prog); err != nil {
			return nil, fmt.Errorf("unmarshal: %w", err)
		}

		if c, err = prog.compile(); err != nil {
			return nil, err
		}
	}

	bs, err := c.encode()
	if err != nil {
		return nil, fmt.Errorf("encode: %w", err)
	}

	return apiprogram.NewModule(
		l.name,
		nil,
		compilerVersion,
		path,
		bs,
	)
}

func (l *langvalues) registerModule(ctx context.Context, env *lang.RunEnv, bind *apivalues.Value, m *compiledModule) error {
	vs, _, err := env.Load(ctx, m.Path)
	if err != nil {
		return fmt.Errorf("load %q: %w", m.Path, err)
	}

	name := m.ValueName
	if name == "" {
		name = "setup"
	}

	specv := vs[name]
	if specv == nil {
		return fmt.Errorf("%q: no such value %q", m.Path, name)
	}

	specl := apivalues.GetListValue(specv)
	if specl == nil {
		return fmt.Errorf("%q: spec is not a list", m.Path)
	}

	for i, mv := range specl {
		mvl := apivalues.GetListValue(mv)
		if mvl == nil {
			return fmt.Errorf("%q: %d: expected mapping spec", m.Path, i)
		}

		if len(mvl) < 2 {
			return fmt.Errorf("%q: %d: mapping must have at least three items", m.Path, i)
		}

		for _, rv := range mvl[2:] {
			if m.Context[rv.String()] == nil {
				return fmt.Errorf("%q: %d: require context value %q not supplied", m.Path, i, rv)
			}
		}

		src := mvl[0].String()
		src1, rest, ok := strings.Cut(src, ".")
		if !ok {
			return fmt.Errorf("%q: %d: invalid source name", m.Path, i)
		}

		src2 := m.Sources[src1]
		if src2 == "" {
			return fmt.Errorf("%q: %d: no mapping for source %q", m.Path, i, src1)
		}

		src = fmt.Sprintf("%s.%s", src2, rest)

		_, err := env.Call(ctx, bind, map[string]*apivalues.Value{
			"source":  apivalues.String(src),
			"target":  mvl[1],
			"context": apivalues.DictFromMap(m.Context),
		}, nil, nil)

		if err != nil {
			return fmt.Errorf("bind: %w", err)
		}
	}

	return nil
}

func (l *langvalues) RunModule(
	ctx context.Context,
	env *lang.RunEnv,
	mod *apiprogram.Module, // mod must have compiled_code populated.
) (map[string]*apivalues.Value, *apilang.RunSummary, error) {
	if cv := mod.CompilerVersion(); cv != compilerVersion {
		return nil, nil, fmt.Errorf("compiler version mismatch, %s != supported %s", cv, compilerVersion)
	}

	var compiled compiled

	if err := json.Unmarshal(mod.CompiledCode(), &compiled); err != nil {
		return nil, nil, fmt.Errorf("proto unmarshal: %w", err)
	}

	if compiled.Consts == nil {
		compiled.Consts = map[string]*apivalues.Value{}
	}

	if l.dataOnly {
		return compiled.Consts, nil, nil
	}

	// TODO: make sure this is only possible in the main file.
	akvs, _, err := env.Load(ctx, apiprogram.MustParsePathString("internal.ak"))
	if err != nil {
		return nil, nil, err
	}

	aksrcsv := akvs["sources"]
	if aksrcsv == nil {
		return nil, nil, fmt.Errorf("no sources submodule in ak")
	}

	sources := apivalues.GetStructValue(aksrcsv)
	if sources == nil {
		return nil, nil, fmt.Errorf("sources is not a struct in ak")
	}

	bind := sources.Fields["bind"]
	if bind == nil {
		return nil, nil, fmt.Errorf("no bind member in ak.sources")
	}

	for _, m := range compiled.Modules {
		if err := l.registerModule(ctx, env, bind, m); err != nil {
			return nil, nil, err
		}
	}

	return compiled.Consts, nil, nil
}

func (*langvalues) CallFunction(context.Context, *lang.RunEnv, *apivalues.Value, []*apivalues.Value, map[string]*apivalues.Value) (*apivalues.Value, *apilang.RunSummary, error) {
	return nil, nil, fmt.Errorf("not supported")
}
