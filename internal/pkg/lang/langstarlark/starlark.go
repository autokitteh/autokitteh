package langstarlark

import (
	"bytes"
	"context"
	"fmt"
	"sort"

	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"

	"github.com/autokitteh/autokitteh/internal/pkg/lang"
	"github.com/autokitteh/autokitteh/internal/pkg/lang/langtools"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apilang"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiprogram"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apivalues"
	L "github.com/autokitteh/autokitteh/pkg/l"
)

type langstarlark struct {
	l        L.Nullable
	name     string
	vs       *values
	builtins map[string]starlark.Value

	// TODO: this might inadveredly leak data between invocations?
	//       that is if a global state is managed by the module and reused
	//       between loads.
	modules map[string]func() (starlark.StringDict, error)
}

var compilerVersion = fmt.Sprintf("%d", starlark.CompilerVersion)

var (
	DeterministicCatalogLang = lang.CatalogLang{New: NewDeterministic, Exts: []string{"kitteh"}}
	PermissiveCatalogLang    = lang.CatalogLang{New: NewPermissive, Exts: []string{"starlark", "star"}}
)

func init() {
	langtools.PermissiveCatalog.Register("starlark", PermissiveCatalogLang)
	langtools.PermissiveCatalog.Register("kitteh", DeterministicCatalogLang)

	langtools.DeterministicCatalog.Register("kitteh", DeterministicCatalogLang)

	resolve.AllowSet = true
	resolve.AllowRecursion = true
}

func NewPermissive(l L.L, name string) (lang.Lang, error) {
	lng := &langstarlark{
		l:        L.N(l),
		name:     name,
		builtins: PermissiveBuiltinValues,
		modules:  PermissiveBuiltinModules,
	}
	lng.Reset()
	return lng, nil
}

func NewDeterministic(l L.L, name string) (lang.Lang, error) {
	lng := &langstarlark{
		l:        L.N(l),
		name:     name,
		builtins: DeterministicBuiltinValues,
		modules:  DeterministicBuiltinModules,
	}
	lng.Reset()
	return lng, nil
}

func (s *langstarlark) Reset() { s.vs = newValues(s) }

func (s *langstarlark) IsCompilerVersionSupported(_ context.Context, v string) (bool, error) {
	return v == compilerVersion, nil
}

func (s *langstarlark) CompileModule(
	_ context.Context,
	path *apiprogram.Path,
	src []byte,
	predecls []string,
) (*apiprogram.Module, error) {
	for k := range s.builtins {
		predecls = append(predecls, k)
	}

	sort.Strings(predecls)

	_, mod, err := starlark.SourceProgram(path.Path(), src, func(n string) bool {
		i := sort.SearchStrings(predecls, n)
		return i < len(predecls) && predecls[i] == n
	})

	if err != nil {
		return nil, errf("compile: %w", err)
	}

	var compiled bytes.Buffer
	if err := mod.Write(&compiled); err != nil {
		return nil, errf("write: %w", err)
	}

	return apiprogram.NewModule(s.name, predecls, compilerVersion, path, compiled.Bytes())
}

func (s *langstarlark) GetModuleDependencies(ctx context.Context, mod *apiprogram.Module) (paths []*apiprogram.Path, _ error) {
	if supported, _ := s.IsCompilerVersionSupported(ctx, mod.CompilerVersion()); !supported {
		return nil, errf("module compiler version (%s) is not supported, expected %s", mod.CompilerVersion(), compilerVersion)
	}

	compiled, err := starlark.CompiledProgram(bytes.NewReader(mod.CompiledCode()))
	if err != nil {
		return nil, errf("decode: %w", err)
	}

	for l := 0; l < compiled.NumLoads(); l++ {
		load, _ := compiled.Load(l)

		path, err := apiprogram.ParsePathString(load)
		if err != nil {
			return nil, errf("path %q: %w", load, err)
		}

		// only report non-builtins.
		if s.modules[path.Path()] == nil {
			paths = append(paths, path)
		}
	}

	return
}

func filterGlobals(gs map[string]starlark.Value) map[string]starlark.Value {
	rs := make(map[string]starlark.Value, len(gs))
	for k, v := range gs {
		if len(k) == 0 || k[0] == '_' {
			continue
		}

		rs[k] = v
	}
	return rs
}

func (s *langstarlark) toStarlarkPredecls(env *lang.RunEnv) (map[string]starlark.Value, error) {
	lvs := s.vs.WithEnv(env)

	predecls, err := lvs.ToStringDict(env.Predecls)
	if err != nil {
		return nil, errf("predecls: %w", err)
	}

	// TODO: std to be overridden by predecls?

	for k, v := range s.builtins {
		if _, ok := predecls[k]; ok {
			return nil, errf("%v is already in use", k)
		}

		predecls[k] = v
	}

	return predecls, nil
}

func (s *langstarlark) RunModule(
	ctx context.Context,
	env *lang.RunEnv,
	mod *apiprogram.Module,
) (map[string]*apivalues.Value, *apilang.RunSummary, error) {
	env = env.WithStubs()

	// init predecls

	predecls, err := s.toStarlarkPredecls(env)
	if err != nil {
		return nil, nil, err
	}

	s.l.Debug("predecls", "predecls", predecls)

	// init thread and loader

	thread := &starlark.Thread{
		Name:  mod.SourcePath().String(),
		Print: func(_ *starlark.Thread, msg string) { env.Print(msg) },
		Load: func(_ *starlark.Thread, module string) (starlark.StringDict, error) {
			if s.modules != nil {
				if load, found := s.modules[module]; found {
					return load()
				}
			}

			p, err := apiprogram.ParsePathString(module)
			if err != nil {
				return nil, errf("new path %q: %w", module, err)
			}

			vs, _, err := env.Load(ctx, p)
			if err != nil {
				return nil, errf("load %q: %w", p, err)
			}

			if vs == nil {
				return nil, errf("load %q: not found", p)
			}

			ret, err := s.vs.WithEnv(env).ToStringDict(vs)
			if err != nil {
				return nil, errf("convert result: %w", err)
			}

			ret = filterGlobals(ret)

			// Allow to explicitly specify exports.
			if _, ok := ret["exports"]; !ok {
				members := make(map[string]starlark.Value, len(ret))
				for k, v := range ret {
					members[k] = v
				}

				exports := &starlarkstruct.Module{Name: module, Members: members}
				ret["exports"] = exports
			}

			// Allow to import nothing (just execute side effects).
			if _, ok := ret["none"]; !ok {
				ret["none"] = starlark.None
			}

			return ret, nil
		},
	}

	setTLSContext(thread, ctx)
	setTLSEnv(thread, env)

	// decode program

	p, err := starlark.CompiledProgram(bytes.NewReader(mod.CompiledCode()))
	if err != nil {
		return nil, nil, errf("code decode: %w", err)
	}

	// run program in separate goroutine (this will enable cancellation)

	type ret struct {
		g   starlark.StringDict
		err error
	}

	ch := make(chan ret, 1)

	go func() {
		g, err := p.Init(thread, predecls)
		g.Freeze()

		ch <- ret{g: g, err: err}
	}()

	// wait for program to be done or canceled

	var g starlark.StringDict

	select {
	case <-ctx.Done():
		thread.Cancel(ctx.Err().Error())
		r := <-ch
		g, err = r.g, r.err

	case r := <-ch:
		g, err = r.g, translateError(r.err)
	}

	if err != nil {
		return nil, nil, translateError(err)
	}

	// return globals

	g = filterGlobals(g)

	gvs, err := s.vs.WithEnv(env).FromStringDict(g, nil)
	if err != nil {
		return nil, nil, errf("convert: %w", err)
	}

	return gvs, nil, nil
}

func (s *langstarlark) CallFunction(
	ctx context.Context,
	env *lang.RunEnv,
	v *apivalues.Value,
	args []*apivalues.Value,
	kws map[string]*apivalues.Value,
) (*apivalues.Value, *apilang.RunSummary, error) {
	fn, ok := v.Get().(apivalues.FunctionValue)
	if !ok {
		return nil, nil, errf("value is not a function")
	}

	env = env.WithStubs()
	lvs := s.vs.WithEnv(env)

	slfn, err := lvs.retreiveFunc(fn)
	if err != nil {
		return nil, nil, translateError(err)
	}

	slargs, err := lvs.tos(args)
	if err != nil {
		return nil, nil, errf("convert args: %w", err)
	}

	slkws, err := lvs.ToStringDict(kws)
	if err != nil {
		return nil, nil, errf("convert kwargs: %w", err)
	}

	slkwstup := make([]starlark.Tuple, 0, len(slkws))
	for k, v := range slkws {
		slkwstup = append(slkwstup, []starlark.Value{starlark.String(k), v})
	}

	// init predecls

	predecls, err := s.toStarlarkPredecls(env)
	if err != nil {
		return nil, nil, err
	}

	s.l.Debug("predecls", "predecls", predecls)

	// init thread and loader

	thread := &starlark.Thread{
		Name:  fn.FuncID,
		Print: func(_ *starlark.Thread, msg string) { env.Print(msg) },
		Load: func(*starlark.Thread, string) (starlark.StringDict, error) {
			return nil, errf("load not supported during a function call")
		},
	}

	setTLSContext(thread, ctx)

	// run function in separate goroutine (this will enable cancellation)

	type ret struct {
		v   starlark.Value
		err error
	}

	ch := make(chan ret, 1)

	go func() {
		v, err := starlark.Call(thread, slfn, slargs, slkwstup)
		ch <- ret{v: v, err: err}
	}()

	// wait for the function to be done or canceled

	var retv ret

	select {
	case <-ctx.Done():
		thread.Cancel(ctx.Err().Error())
		retv = <-ch

	case retv = <-ch:
		retv.err = translateError(retv.err)
	}

	if retv.err != nil {
		return nil, nil, translateError(retv.err)
	}

	gv, err := s.vs.WithEnv(env).FromStarlarkValue(retv.v, nil)
	if err != nil {
		return nil, nil, errf("convert: %w", err)
	}

	return gv, nil, nil
}
