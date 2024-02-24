package kittehs

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/psanford/memfs"
)

func FSToMap(in fs.FS) (out map[string][]byte, err error) {
	out = make(map[string][]byte)
	err = fs.WalkDir(in, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		bs, err := fs.ReadFile(in, path)
		if err != nil {
			return err
		}

		out[path] = bs

		return nil
	})
	return
}

func MapToMemFS(in map[string][]byte) (fs.FS, error) {
	memfs := memfs.New()

	for path, bs := range in {
		path := filepath.Clean(path)

		base := filepath.Dir(path)
		if base != "" {
			if err := memfs.MkdirAll(base, 0o755); err != nil {
				return nil, fmt.Errorf("failed to create directory %s: %w", base, err)
			}
		}

		if err := memfs.WriteFile(path, bs, 0o644); err != nil {
			return nil, fmt.Errorf("failed to write file %s: %w", path, err)
		}
	}

	return memfs, nil
}
