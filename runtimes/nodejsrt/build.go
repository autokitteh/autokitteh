package nodejsrt

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

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func isBuildFile(entry fs.DirEntry) bool {
	// Skip node_modules, dist directory and hidden files
	if strings.Contains(entry.Name(), "node_modules") || strings.Contains(entry.Name(), "dist") {
		return false
	}

	if path.Base(entry.Name())[0] == '.' {
		return false
	}

	return true
}

// TODO: Move build to runner
// TODO: optional, build manager like remote runner vs local runner
func (js *nodejsSvc) Build(ctx context.Context, fsys fs.FS, path string, values []sdktypes.Symbol) (sdktypes.BuildArtifact, error) {
	js.log.Info("build NodeJS module", zap.String("path", path))

	// 1. Create temp dir for output
	outputDir, err := os.MkdirTemp("", "ak-proj")
	if err != nil {
		return sdktypes.InvalidBuildArtifact, err
	}
	defer os.RemoveAll(outputDir)

	// 2. Create temp dir for input
	inputDir, err := os.MkdirTemp("", "ak-proj-input")
	if err != nil {
		return sdktypes.InvalidBuildArtifact, err
	}
	defer os.RemoveAll(inputDir)

	// Copy input files to input dir using copyFSToDir
	if err := copyFSToDir(fsys, inputDir); err != nil {
		return sdktypes.InvalidBuildArtifact, err
	}

	// 3. Create temp dir for runner tools
	runnerDir, err := os.MkdirTemp("", "ak-runner")
	if err != nil {
		return sdktypes.InvalidBuildArtifact, err
	}
	defer os.RemoveAll(runnerDir)

	// Copy runner tools using copyFSToDir
	if err := copyFSToDir(runnerJsCode, runnerDir); err != nil {
		return sdktypes.InvalidBuildArtifact, err
	}

	// 4. Run build.ts to copy and modify the code
	buildCmd := exec.Command("node", "runner/node_modules/ts-node/dist/bin.js", "runner/build.ts", inputDir, outputDir)
	buildCmd.Dir = runnerDir
	var buildStdout, buildStderr bytes.Buffer
	buildCmd.Stdout = &buildStdout
	buildCmd.Stderr = &buildStderr

	if err := buildCmd.Run(); err != nil {
		js.log.Error("build.ts failed",
			zap.Error(err),
			zap.String("stdout", buildStdout.String()),
			zap.String("stderr", buildStderr.String()))
		return sdktypes.InvalidBuildArtifact, fmt.Errorf("build.ts failed:\n%s", buildStderr.String())
	}

	// 5. Create tar from the modified files
	data, err := createTar(os.DirFS(outputDir))
	if err != nil {
		js.log.Error("create tar", zap.Error(err))
		return sdktypes.InvalidBuildArtifact, err
	}

	compiledData := map[string][]byte{
		archiveKey: data,
	}

	// Add file names to compiledData for UI
	tf := tar.NewReader(bytes.NewReader(data))
	for {
		hdr, err := tf.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			js.log.Error("next tar", zap.Error(err))
			return sdktypes.InvalidBuildArtifact, err
		}

		if !strings.HasSuffix(hdr.Name, ".js") && !strings.HasSuffix(hdr.Name, ".ts") {
			continue
		}

		compiledData[hdr.Name] = nil
	}

	// 6. Find exports from the modified code
	exports, err := findExports(js.log, os.DirFS(outputDir))
	if err != nil {
		js.log.Error("get exports", zap.Error(err))
		return sdktypes.InvalidBuildArtifact, err
	}

	var art sdktypes.BuildArtifact
	art = art.WithCompiledData(compiledData)
	art = art.WithExports(exports)
	return art, nil
}

func findExports(log *zap.Logger, fsys fs.FS) ([]sdktypes.BuildExport, error) {
	codeDir, err := os.MkdirTemp("", "ak-proj")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(codeDir)

	if err := copyFSToDir(fsys, codeDir); err != nil {
		return nil, err
	}

	runnerDir, err := os.MkdirTemp("", "ak-runner")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(runnerDir)

	if err := copyFSToDir(runnerJsCode, runnerDir); err != nil {
		return nil, err
	}

	// Run exports discovery
	log.Info("discovering exports")
	cmd := exec.Command("node", "runner/node_modules/ts-node/dist/bin.js", "runner/exports.ts", codeDir)
	cmd.Dir = runnerDir
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
