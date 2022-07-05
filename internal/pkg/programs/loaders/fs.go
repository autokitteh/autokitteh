package loaders

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"go.autokitteh.dev/sdk/api/apiprogram"
)

func NewRootFSLoader() LoaderFunc {
	return NewFSLoader(nil, "")
}

// WARNING: does not protect again relative paths.
// [# path-protect #] should do that.
func NewFSLoader(rootFS fs.FS, root string) LoaderFunc {
	return func(_ context.Context, given *apiprogram.Path) ([]byte, string, error) {
		if given.Version() != "" {
			return nil, "", fmt.Errorf("fs loader does not support versions")
		}

		var (
			src []byte
			err error
		)

		if rootFS == nil {
			if root != "" {
				return nil, "", fmt.Errorf("if not fs is given, root must be empty")
			}

			src, err = os.ReadFile(given.Path())
		} else {
			path := filepath.Join(root, given.Path())

			var rel string
			if rel, err = filepath.Rel(root, path); err != nil {
				return nil, "", fmt.Errorf("%q is not relative to %q: %w", path, root, err)
			}

			if rel[0] == '.' {
				return nil, "", fmt.Errorf("%q cannot escape %q", path, root)
			}

			src, err = fs.ReadFile(rootFS, path)
		}

		if err != nil {
			if os.IsNotExist(err) {
				return nil, "", ErrNotFound
			}

			return nil, "", fmt.Errorf("read: %w", err)
		}

		return src, fmt.Sprintf("%x", sha256.Sum256(src)), nil
	}
}
