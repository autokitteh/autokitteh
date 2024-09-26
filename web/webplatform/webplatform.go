package webplatform

import (
	"archive/zip"
	"bytes"
	"embed"
	"errors"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/kittehs"

	"github.com/psanford/memfs"
)

//go:embed autokitteh-web-*.zip
var zipFS embed.FS

// Loads the first zip file found in the embedded filesystem and
// extracts the content under its dist/ directory in a memory filesystem.
func LoadFS() (fs.FS, string, error) {
	des, err := fs.ReadDir(zipFS, ".")
	if err != nil {
		return nil, "", err
	}

	_, de := kittehs.FindFirst(des, func(de fs.DirEntry) bool {
		return de.Type().IsRegular() && strings.HasSuffix(de.Name(), ".zip")
	})

	if de == nil {
		return nil, "", errors.New("no zip file found")
	}

	zipFilename := de.Name()

	memfs := memfs.New()

	version := strings.TrimSuffix(zipFilename, ".zip")

	data, err := fs.ReadFile(zipFS, zipFilename)
	if err != nil {
		return nil, version, err
	}

	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, version, err
	}

	for _, f := range r.File {
		name := f.Name

		if f.FileInfo().IsDir() || !strings.HasPrefix(name, "dist/") {
			continue
		}

		name = strings.TrimPrefix(name, "dist/")

		fr, err := f.Open()
		if err != nil {
			return nil, version, err
		}

		data, err := io.ReadAll(fr)
		if err != nil {
			fr.Close()
			return nil, version, err
		}

		fr.Close()

		if err := memfs.MkdirAll(filepath.Dir(name), 0o755); err != nil {
			return nil, version, err
		}

		if err := memfs.WriteFile(name, data, 0o644); err != nil {
			return nil, version, err
		}
	}

	return memfs, version, nil
}
