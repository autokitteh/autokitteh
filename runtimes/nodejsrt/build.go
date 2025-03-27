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
	"path/filepath"
	"runtime"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func runNodeScript(log *zap.Logger, scriptPath string, args ...string) ([]byte, []byte, error) {
	log.Info("executing Node.js script",
		zap.String("scriptPath", scriptPath),
		zap.Strings("args", args),
	)
	_, currentFile, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(currentFile)
	nodeDir := filepath.Join(baseDir, "nodejs") // adjust as needed

	cmd := exec.Command("npx", append([]string{"ts-node", scriptPath}, args...)...)
	cmd.Dir = nodeDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, nil, fmt.Errorf("script execution failed:\n%s", stderr.String())
	}

	return stdout.Bytes(), stderr.Bytes(), nil
}

// TODO: Move build to runner
// TODO: optional, build manager like remote runner vs local runner
func (js *nodejsSvc) Build(ctx context.Context, fsys fs.FS, path string, values []sdktypes.Symbol) (sdktypes.BuildArtifact, error) {
	js.log.Info("build NodeJS module", zap.String("path", path))

	// 1. Create temp dir for input and copy
	inputDir, err := os.MkdirTemp("", "ak-proj-input")
	if err != nil {
		return sdktypes.InvalidBuildArtifact, err
	}
	defer os.RemoveAll(inputDir)
	if err := copyFSToDir(fsys, inputDir); err != nil {
		return sdktypes.InvalidBuildArtifact, err
	}

	// 2. Create temp dir for output
	outputDir, err := os.MkdirTemp("", "ak-proj")
	if err != nil {
		return sdktypes.InvalidBuildArtifact, err
	}
	defer os.RemoveAll(outputDir)

	// 3. Run build.ts to copy and modify the code
	buildStdout, buildStderr, err := runNodeScript(js.log, "builder/build.ts", inputDir, outputDir)
	if err != nil {
		js.log.Error("build.ts failed",
			zap.Error(err),
			zap.String("stdout", string(buildStdout)),
			zap.String("stderr", string(buildStderr)))
		return sdktypes.InvalidBuildArtifact, fmt.Errorf("build.ts failed:\n%s", string(buildStderr))
	}

	// 4. Create tar from the modified files
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

	// 5. Find exports from the modified code
	exports, err := findExports(js.log, outputDir)
	if err != nil {
		js.log.Error("get exports", zap.Error(err))
		return sdktypes.InvalidBuildArtifact, err
	}

	var art sdktypes.BuildArtifact
	art = art.WithCompiledData(compiledData)
	art = art.WithExports(exports)
	return art, nil
}

func findExports(log *zap.Logger, codeDir string) ([]sdktypes.BuildExport, error) {
	// Run exports discovery
	stdout, stderr, err := runNodeScript(log, "builder/exports.ts", codeDir)
	log.Info("discovering exports")
	if err != nil {
		return nil, fmt.Errorf("export:\n%s", stderr)
	}

	var exports []Export
	if err := json.Unmarshal(stdout, &exports); err != nil {
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
