package langbin

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"go.autokitteh.dev/sdk/api/apilang"
	"go.autokitteh.dev/sdk/api/apiprogram"
	"go.autokitteh.dev/sdk/api/apivalues"

	"github.com/autokitteh/L"
	"github.com/autokitteh/autokitteh/internal/pkg/lang"
	"github.com/autokitteh/autokitteh/internal/pkg/lang/langtools"
)

type DecoderFn func([]byte) ([]byte, error)

type langbin struct {
	name    string
	decoder DecoderFn
}

var compilerVersion = "0"

func init() {
	Register(langtools.PermissiveCatalog)
	Register(langtools.DeterministicCatalog)
}

func Register(cat lang.Catalog) {
	cat.Register("bin", lang.CatalogLang{
		New:  NewBinLang(func(src []byte) ([]byte, error) { return src, nil }),
		Exts: []string{"bin"},
	})
	cat.Register("base64", lang.CatalogLang{
		New: NewBinLang(func(src []byte) ([]byte, error) {
			src = bytes.TrimSpace(src)
			dst := make([]byte, base64.StdEncoding.EncodedLen(len(src)))
			n, err := base64.StdEncoding.Decode(dst, src)
			if err != nil {
				return nil, err
			}
			return dst[:n], nil
		}),
		Exts: []string{"base64"},
	})
	cat.Register("hex", lang.CatalogLang{
		New: NewBinLang(func(src []byte) ([]byte, error) {
			src = bytes.TrimSpace(src)
			dst := make([]byte, hex.EncodedLen(len(src)))
			n, err := hex.Decode(dst, src)
			if err != nil {
				return nil, err
			}
			return dst[:n], nil
		}),
		Exts: []string{"hex"},
	})
}

func NewBinLang(decoder DecoderFn) func(L.L, string) (lang.Lang, error) {
	return func(_ L.L, name string) (lang.Lang, error) {
		return &langbin{name: name, decoder: decoder}, nil
	}
}

func (*langbin) IsCompilerVersionSupported(_ context.Context, v string) (bool, error) {
	return v == compilerVersion, nil
}

func (*langbin) GetModuleDependencies(_ context.Context, mod *apiprogram.Module) ([]*apiprogram.Path, error) {
	return nil, nil
}

func (l *langbin) CompileModule(
	ctx context.Context,
	path *apiprogram.Path,
	src []byte,
	_ []string,
) (*apiprogram.Module, error) {
	data, err := l.decoder(src)
	if err != nil {
		return nil, fmt.Errorf("decoder error: %w", err)
	}

	return apiprogram.NewModule(
		l.name,
		nil,
		compilerVersion,
		path,
		data,
	)
}

func (l *langbin) RunModule(
	ctx context.Context,
	env *lang.RunEnv,
	mod *apiprogram.Module, // mod must have compiled_code populated.
) (map[string]*apivalues.Value, *apilang.RunSummary, error) {
	if cv := mod.CompilerVersion(); cv != compilerVersion {
		return nil, nil, fmt.Errorf("compiler version mismatch, %s != supported %s", cv, compilerVersion)
	}

	return map[string]*apivalues.Value{
		"bytes": apivalues.Bytes(mod.CompiledCode()),
	}, nil, nil
}

func (*langbin) CallFunction(context.Context, *lang.RunEnv, *apivalues.Value, []*apivalues.Value, map[string]*apivalues.Value) (*apivalues.Value, *apilang.RunSummary, error) {
	return nil, nil, fmt.Errorf("not supported")
}
