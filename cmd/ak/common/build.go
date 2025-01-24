package common

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkbuildfile"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func Build(rts sdkservices.Runtimes, srcFS fs.FS, paths []string, syms []sdktypes.Symbol) (*sdkbuildfile.BuildFile, error) {
	ctx, cancel := LimitedContext()
	defer cancel()

	if len(paths) != 0 {
		files := make(map[string][]byte, len(paths))

		for _, path := range paths {
			data, err := fs.ReadFile(srcFS, filepath.Clean(path))
			if err != nil {
				return nil, fmt.Errorf("read file %q: %w", path, err)
			}

			files[path] = data
		}

		var err error
		if srcFS, err = kittehs.MapToMemFS(files); err != nil {
			return nil, fmt.Errorf("create memory filesystem: %w", err)
		}
	}

	b, err := rts.Build(ctx, srcFS, syms, nil)
	if err != nil {
		return nil, fmt.Errorf("create build: %w", err)
	}

	return b, nil
}

func CollectUploads(dirPaths, filePaths []string) (map[string][]byte, error) {
	uploads := make(map[string][]byte)
	for _, path := range append(dirPaths, filePaths...) {
		fi, err := os.Stat(path)
		if err != nil {
			return nil, NewExitCodeError(NotFoundExitCode, err)
		}

		// Upload an entire directory tree.
		if fi.IsDir() {
			err := filepath.WalkDir(path, walk(path, uploads))
			if err != nil {
				return nil, err
			}
			continue
		}

		// Upload a single file.
		contents, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		uploads[fi.Name()] = contents
	}

	return uploads, nil
}

func BuildProject(pid sdktypes.ProjectID, dirPaths, filePaths []string) (sdktypes.BuildID, error) {
	ctx, cancel := LimitedContext()
	defer cancel()

	uploads, err := CollectUploads(dirPaths, filePaths)
	if err != nil {
		return sdktypes.InvalidBuildID, err
	}

	// Communicate with the server in 2 steps.
	if err := Client().Projects().SetResources(ctx, pid, uploads); err != nil {
		return sdktypes.InvalidBuildID, fmt.Errorf("set resources: %w", err)
	}

	bid, err := Client().Projects().Build(ctx, pid)
	if err != nil {
		return sdktypes.InvalidBuildID, fmt.Errorf("build project: %w", err)
	}

	return bid, nil
}

func walk(basePath string, uploads map[string][]byte) fs.WalkDirFunc {
	return func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err // Abort the entire walk.
		}
		if d.IsDir() {
			return nil // Skip directory analysis, focus on files.
		}

		// Upload a single file, relative to the base path.
		relPath, err := filepath.Rel(basePath, path)
		if err != nil {
			return fmt.Errorf("relative path: %w", err)
		}

		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		uploads[relPath] = contents
		return nil
	}
}
