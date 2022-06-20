package langtxtar

import (
	"context"
	"fmt"

	"golang.org/x/tools/txtar"

	"go.autokitteh.dev/sdk/api/apilang"
	"go.autokitteh.dev/sdk/api/apiprogram"
	"go.autokitteh.dev/sdk/api/apivalues"

	"github.com/autokitteh/L"
	"github.com/autokitteh/autokitteh/internal/pkg/lang"
	"github.com/autokitteh/autokitteh/internal/pkg/lang/langtools"
)

type langtxtar struct {
	name string
}

var compilerVersion = "0"

func init() {
	Register(langtools.PermissiveCatalog)
	Register(langtools.DeterministicCatalog)
}

func Register(cat lang.Catalog) {
	cat.Register("json-program", lang.CatalogLang{
		New:  NewTxtarLang,
		Exts: []string{"kitteh.json"},
	})
}

func NewTxtarLang(_ L.L, name string) (lang.Lang, error) {
	return &langtxtar{name: name}, nil
}

func (*langtxtar) IsCompilerVersionSupported(_ context.Context, v string) (bool, error) {
	return v == compilerVersion, nil
}

func (*langtxtar) GetModuleDependencies(_ context.Context, mod *apiprogram.Module) ([]*apiprogram.Path, error) {
	return nil, nil
}

func (l *langtxtar) CompileModule(
	ctx context.Context,
	path *apiprogram.Path,
	src []byte,
	_ []string,
) (*apiprogram.Module, error) {
	return apiprogram.NewModule(
		l.name,
		nil,
		compilerVersion,
		path,
		src,
	)
}

func (l *langtxtar) RunModule(
	ctx context.Context,
	env *lang.RunEnv,
	mod *apiprogram.Module, // mod must have compiled_code populated.
) (map[string]*apivalues.Value, *apilang.RunSummary, error) {
	if cv := mod.CompilerVersion(); cv != compilerVersion {
		return nil, nil, fmt.Errorf("compiler version mismatch, %s != supported %s", cv, compilerVersion)
	}

	arch := txtar.Parse(mod.CompiledCode())

	fs := make(map[string]*apivalues.Value, len(arch.Files))
	for _, f := range arch.Files {
		fs[f.Name] = apivalues.String(string(f.Data))
	}

	return map[string]*apivalues.Value{
		"comment": apivalues.String(string(arch.Comment)),
		"files":   apivalues.DictFromMap(fs),
	}, nil, nil
}

func (*langtxtar) CallFunction(context.Context, *lang.RunEnv, *apivalues.Value, []*apivalues.Value, map[string]*apivalues.Value) (*apivalues.Value, *apilang.RunSummary, error) {
	return nil, nil, fmt.Errorf("not supported")
}
