package langcsv

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"

	"google.golang.org/protobuf/proto"

	pbvalues "go.autokitteh.dev/idl/go/values"
	"go.autokitteh.dev/sdk/api/apilang"
	"go.autokitteh.dev/sdk/api/apiprogram"
	"go.autokitteh.dev/sdk/api/apivalues"

	"github.com/autokitteh/L"
	"github.com/autokitteh/autokitteh/internal/pkg/lang"
	"github.com/autokitteh/autokitteh/internal/pkg/lang/langtools"
)

type langcsv struct {
	name string
}

var compilerVersion = "0"

func init() {
	Register(langtools.PermissiveCatalog)
	Register(langtools.DeterministicCatalog)
}

func Register(cat lang.Catalog) {
	cat.Register("csv", lang.CatalogLang{
		New:  NewCSVLang,
		Exts: []string{"csv"},
	})
}

func NewCSVLang(_ L.L, name string) (lang.Lang, error) {
	return &langcsv{name: name}, nil
}

func (*langcsv) IsCompilerVersionSupported(_ context.Context, v string) (bool, error) {
	return v == compilerVersion, nil
}

func (*langcsv) GetModuleDependencies(_ context.Context, mod *apiprogram.Module) ([]*apiprogram.Path, error) {
	return nil, nil
}

func (l *langcsv) CompileModule(
	ctx context.Context,
	path *apiprogram.Path,
	src []byte,
	_ []string,
) (*apiprogram.Module, error) {
	rs, err := csv.NewReader(bytes.NewReader(src)).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("ReadAll: %w", err)
	}

	lvs := make([]*apivalues.Value, len(rs))
	for i, r := range rs {
		vs := make([]*apivalues.Value, len(r))
		for j, v := range r {
			vs[j] = apivalues.String(v)
		}
		lvs[i] = apivalues.List(vs...)
	}

	pb := apivalues.List(lvs...).PB()
	data, err := proto.Marshal(pb)
	if err != nil {
		return nil, fmt.Errorf("proto marshal: %w", err)
	}

	return apiprogram.NewModule(
		l.name,
		nil,
		compilerVersion,
		path,
		data,
	)
}

func (l *langcsv) RunModule(
	ctx context.Context,
	env *lang.RunEnv,
	mod *apiprogram.Module, // mod must have compiled_code populated.
) (map[string]*apivalues.Value, *apilang.RunSummary, error) {
	if cv := mod.CompilerVersion(); cv != compilerVersion {
		return nil, nil, fmt.Errorf("compiler version mismatch, %s != supported %s", cv, compilerVersion)
	}

	var pb pbvalues.Value
	if err := proto.Unmarshal(mod.CompiledCode(), &pb); err != nil {
		return nil, nil, fmt.Errorf("unmarshal: %w", err)
	}

	v, err := apivalues.ValueFromProto(&pb)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value: %w", err)
	}

	return map[string]*apivalues.Value{
		"records": v,
	}, nil, nil
}

func (*langcsv) CallFunction(context.Context, *lang.RunEnv, *apivalues.Value, []*apivalues.Value, map[string]*apivalues.Value) (*apivalues.Value, *apilang.RunSummary, error) {
	return nil, nil, fmt.Errorf("not supported")
}
