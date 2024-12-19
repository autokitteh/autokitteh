package pythonrt

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.uber.org/zap"
)

func isBuildFile(entry fs.DirEntry) bool {
	if strings.Contains(entry.Name(), "__pycache__") {
		return false
	}

	if path.Base(entry.Name())[0] == '.' {
		return false
	}

	return true
}

// TODO: Move build to runner
// TODO: optional, build manager like remote runner vs local runner
func (py *pySvc) Build(ctx context.Context, fsys fs.FS, path string, values []sdktypes.Symbol) (sdktypes.BuildArtifact, error) {
	py.log.Info("build Python module", zap.String("path", path))

	ffs, err := kittehs.NewFilterFS(fsys, isBuildFile)
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

	exports, err := findExports(py.log, fsys)
	if err != nil {
		py.log.Error("get exports", zap.Error(err))
		return sdktypes.InvalidBuildArtifact, err
	}

	art = art.WithExports(exports)
	return art, nil
}

func findExports(log *zap.Logger, fsys fs.FS) ([]sdktypes.BuildExport, error) {
	// TODO(ENG-1784): This has common code with configureLocalRunnerManager, find a way to unite
	sysPyExe, isUserPy, err := pythonToRun(log)
	if err != nil {
		return nil, err
	}

	var pyExe string
	if !isUserPy {
		if err := ensureVEnv(log, sysPyExe); err != nil {
			return nil, err
		}
		pyExe = venvPy
	} else {
		pyExe = sysPyExe
	}

	codeDir, err := os.MkdirTemp("", "ak-proj")
	if err != nil {
		return nil, err
	}

	if err := os.CopyFS(codeDir, fsys); err != nil {
		return nil, err
	}

	runnerDir, err := os.MkdirTemp("", "ak-runner")
	if err != nil {
		return nil, err
	}
	if err := os.CopyFS(runnerDir, runnerPyCode); err != nil {
		return nil, err
	}

	pyFile := path.Join(runnerDir, "runner", "exports.py")
	cmd := exec.Command(pyExe, pyFile, codeDir)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("export:\n%s", stderr.String())
	}

	var exports []Export
	if err := json.Unmarshal(stdout.Bytes(), &exports); err != nil {
		return nil, fmt.Errorf("export: %w", err)
	}

	out := make([]sdktypes.BuildExport, len(exports))
	for i, e := range exports {
		loc, err := sdktypes.CodeLocationFromProto(&sdktypes.CodeLocationPB{
			Path: e.File,
			Col:  e.Line,
			Name: e.Name,
		})
		if err != nil {
			return nil, err
		}

		out[i] = sdktypes.NewBuildExport().WithSymbol(sdktypes.NewSymbol(e.Name)).WithLocation(loc)
	}

	return out, nil
}

type Export struct {
	File string
	Line uint32
	Name string
	Args []string
}
