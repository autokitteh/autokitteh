package langtxt

import (
	"context"
	"fmt"

	"go.autokitteh.dev/sdk/api/apilang"
	"go.autokitteh.dev/sdk/api/apiprogram"
	"go.autokitteh.dev/sdk/api/apivalues"

	"github.com/autokitteh/L"
	"github.com/autokitteh/autokitteh/internal/pkg/lang"
	"github.com/autokitteh/autokitteh/internal/pkg/lang/langtools"
)

type langtxt struct {
	name string
}

var compilerVersion = "0"

func init() {
	Register(langtools.PermissiveCatalog)
	Register(langtools.DeterministicCatalog)
}

func Register(cat lang.Catalog) {
	cat.Register("txt", lang.CatalogLang{
		New:  NewTxtLang,
		Exts: []string{"txt"},
	})
}

func NewTxtLang(_ L.L, name string) (lang.Lang, error) {
	return &langtxt{name: name}, nil
}

func (*langtxt) IsCompilerVersionSupported(_ context.Context, v string) (bool, error) {
	return v == compilerVersion, nil
}

func (*langtxt) GetModuleDependencies(_ context.Context, mod *apiprogram.Module) ([]*apiprogram.Path, error) {
	return nil, nil
}

func (l *langtxt) CompileModule(
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

func (l *langtxt) RunModule(
	ctx context.Context,
	env *lang.RunEnv,
	mod *apiprogram.Module, // mod must have compiled_code populated.
) (map[string]*apivalues.Value, *apilang.RunSummary, error) {
	if cv := mod.CompilerVersion(); cv != compilerVersion {
		return nil, nil, fmt.Errorf("compiler version mismatch, %s != supported %s", cv, compilerVersion)
	}

	return map[string]*apivalues.Value{
		"text": apivalues.String(string(mod.CompiledCode())),
	}, nil, nil
}

func (*langtxt) CallFunction(context.Context, *lang.RunEnv, *apivalues.Value, []*apivalues.Value, map[string]*apivalues.Value) (*apivalues.Value, *apilang.RunSummary, error) {
	return nil, nil, fmt.Errorf("not supported")
}
