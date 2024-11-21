package webplatform

import (
	"archive/zip"
	"bytes"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"

	"github.com/psanford/memfs"
)

//go:embed VERSION *.zip
var zipFS embed.FS

var distRegex = regexp.MustCompile(`^autokitteh-web-v(\d+\.\d+\.\d+)\.zip$`)

func DistFilename() (string, error) {
	des, err := fs.ReadDir(zipFS, ".")
	if err != nil {
		return "", err
	}

	des = kittehs.Filter(des, func(de fs.DirEntry) bool { return distRegex.MatchString(de.Name()) })
	if n := len(des); n == 0 {
		return "", sdkerrors.ErrNotFound
	} else if n != 1 {
		return "", fmt.Errorf("%w: found %d>1 distribution zip files", sdkerrors.ErrConflict, n)
	}

	return des[0].Name(), nil
}

func ensureVersion(l *zap.Logger, loaded string) (bool, error) {
	bs, err := fs.ReadFile(zipFS, "VERSION")
	if err != nil {
		return false, fmt.Errorf("VERSION: %w", sdkerrors.ErrNotFound)
	}

	expected, _, _ := strings.Cut(string(bs), " ")

	if string(expected) != loaded {
		l.Sugar().Warnf("expected webplaftorm VERSION %q != loaded distribution version %q. Run `make ak`?", expected, loaded)
		return false, nil
	}

	return true, nil
}

// Loads the first zip file found in the embedded filesystem and
// extracts the content under its dist/ directory in a memory filesystem.
func LoadFS(l *zap.Logger) (fs.FS, string, error) {
	zipFilename, err := DistFilename()
	if err != nil {
		return nil, "", err
	}

	memfs := memfs.New()

	ms := distRegex.FindAllStringSubmatch(zipFilename, -1)
	version := ms[0][1]

	if _, err := ensureVersion(l, version); err != nil {
		return nil, version, err
	}

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
