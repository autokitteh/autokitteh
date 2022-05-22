package programs

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/autokitteh/autokitteh/sdk/api/apiprogram"
)

// WARNING: does not protect again relative paths.
// [# path-protect #] should do that.
func NewFSLoader(rootFS fs.FS, root string) LoaderFunc {
	return func(_ context.Context, given *apiprogram.Path) ([]byte, error) {
		if given.Version() != "" {
			return nil, fmt.Errorf("fs loader does not support versions")
		}

		path := filepath.Join(root, given.Path())

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil, fmt.Errorf("%q is not relative to %q: %w", path, root, err)
		}

		if rel[0] == '.' {
			return nil, fmt.Errorf("%q cannot escape %q", path, root)
		}

		src, err := fs.ReadFile(rootFS, path)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, ErrNotFound
			}

			return nil, fmt.Errorf("read: %w", err)
		}

		return src, nil
	}
}
