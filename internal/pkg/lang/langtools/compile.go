package langtools

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiprogram"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/lang"
)

var langRe = regexp.MustCompile(`^# =\^\.\^= lang=([0-9A-Za-z_]+)`) // =^.^= (keep for grepping)

func IsCompilerVersionSupported(ctx context.Context, cat lang.Catalog, lang, ver string) (bool, error) {
	l, err := cat.Acquire(lang, "")
	if err != nil {
		return false, fmt.Errorf("lang: %w", err)
	}

	return l.IsCompilerVersionSupported(ctx, ver)
}

func CompileModule(ctx context.Context, cat lang.Catalog, predecls []string, path *apiprogram.Path, src []byte) (_ *apiprogram.Module, ext string, _ error) {
	var name string

	if ms := langRe.FindAllSubmatch(src, -1); len(ms) > 0 {
		name = string(ms[0][1])
	}

	if name == "" {
		var err error
		if name, ext, err = PathToRegisteredLang(cat, path.String()); err != nil {
			return nil, ext, err
		}
	}

	if name == "" {
		return nil, "", errors.New("cannot deduce language")
	}

	lang, err := cat.Acquire(name, "")
	if err != nil {
		return nil, ext, err
	}

	mod, err := lang.CompileModule(ctx, path, src, predecls)

	return mod, ext, err
}
