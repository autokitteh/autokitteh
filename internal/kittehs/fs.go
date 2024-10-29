package kittehs

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/psanford/memfs"
	"golang.org/x/tools/txtar"
)

func FSToMap(in fs.FS) (map[string][]byte, error) {
	out := make(map[string][]byte)
	err := fs.WalkDir(in, ".", func(path string, d fs.DirEntry, err error) error {
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
	if err != nil {
		return nil, err
	}
	return out, nil
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

func TxtarToFS(a *txtar.Archive) (fs.FS, error) {
	m := make(map[string][]byte, len(a.Files))

	for _, f := range a.Files {
		m[f.Name] = f.Data
	}

	return MapToMemFS(m)
}

// FilterFS will filter entries based on a predicate function.
type FilterFS struct {
	fs.FS // embed underlying FS

	Pred func(fs.DirEntry) bool // DirEntry predicate
}

// NewFilterFS returns a new FilterFS wrapping the given FS and predicate.
// `nil` pred assumes to always return true.
func NewFilterFS(fsys fs.FS, pred func(fs.DirEntry) bool) (*FilterFS, error) {
	if fsys == nil {
		return nil, fmt.Errorf("fsys is nil")
	}

	ffs := FilterFS{
		FS:   fsys,
		Pred: pred,
	}

	return &ffs, nil
}

// ReadDir implements fs.ReadDirFS.
func (f *FilterFS) ReadDir(name string) ([]fs.DirEntry, error) {
	entries, err := fs.ReadDir(f.FS, name)
	if err != nil {
		return nil, err
	}

	var filtered []fs.DirEntry
	for _, entry := range entries {
		if f.Pred == nil || f.Pred(entry) {
			filtered = append(filtered, entry)
		}
	}

	return filtered, nil
}
