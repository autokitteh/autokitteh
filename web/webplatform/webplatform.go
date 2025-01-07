package webplatform

import (
	"archive/zip"
	"bytes"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/psanford/memfs"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

//go:embed VERSION *.zip
var distFS embed.FS

// Loads the first zip file found in the embedded filesystem and
// extracts the content under its dist/ directory in a memory filesystem.
func LoadFS(l *zap.Logger) (fs.FS, string, error) {
	bs, err := fs.ReadFile(distFS, "VERSION")
	if err != nil {
		return nil, "", fmt.Errorf("VERSION: %w", err)
	}

	version, _, _ := strings.Cut(string(bs), " ")

	memfs := memfs.New()

	distFilename := fmt.Sprintf("autokitteh-web-v%s.zip", version)

	data, err := fs.ReadFile(distFS, distFilename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", sdkerrors.ErrNotFound
		}

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
