package pythonrt

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.uber.org/zap"
)

// TODO: Move build to runner
// TODO: optional, build manager like remote runner vs local runner
func (py *pySvc) Build(ctx context.Context, fsys fs.FS, path string, values []sdktypes.Symbol) (sdktypes.BuildArtifact, error) {
	py.log.Info("build Python module", zap.String("path", path))

	ffs, err := kittehs.NewFilterFS(fsys, func(entry fs.DirEntry) bool {
		return !strings.Contains(entry.Name(), "__pycache__")
	})
	if err != nil {
		return sdktypes.InvalidBuildArtifact, err
	}

	data, err := createTar(ffs)
	if err != nil {
		py.log.Error("create tar", zap.Error(err))
		return sdktypes.InvalidBuildArtifact, err
	}

	compiledData := map[string][]byte{
		archiveKey: data,
	}

	// UI requires file names in the compiled data.
	tf := tar.NewReader(bytes.NewReader(data))
	for {
		hdr, err := tf.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			py.log.Error("next tar", zap.Error(err))
			return sdktypes.InvalidBuildArtifact, err
		}

		if !strings.HasSuffix(hdr.Name, ".py") {
			continue
		}

		compiledData[hdr.Name] = nil
	}

	var art sdktypes.BuildArtifact
	art = art.WithCompiledData(compiledData)

	// TODO: We don't have exports for now, UI can change singleshot to a text
	// box instead of dropdown

	return art, nil
}
