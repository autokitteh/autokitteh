package common

import (
	"fmt"
	"io/fs"
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
