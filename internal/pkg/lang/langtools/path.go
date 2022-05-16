package langtools

import (
	"strings"

	"github.com/autokitteh/autokitteh/internal/pkg/lang"
)

func SplitExtension(cat lang.Catalog, path string) (base, ext string) {
	_, ext, err := PathToRegisteredLang(cat, path)
	if err != nil {
		return path, ""
	}

	return strings.TrimSuffix(path, "."+ext), ext
}

func PathToRegisteredLang(cat lang.Catalog, path string) (lang_, ext string, err error) {
	// keep track for the longest match (eg .cue vs .kitteh.cue, we want .kitteh.cue).
	var _n, _ext string

	for n, exts := range cat.List() {
		for _, ext := range exts {
			if strings.HasSuffix(path, "."+ext) {
				if len(ext) > len(_ext) {
					_n = n
					_ext = ext
				}
			}
		}
	}

	if _n == "" {
		return "", "", lang.ErrExtensionNotRegistered
	}

	return _n, _ext, nil
}
