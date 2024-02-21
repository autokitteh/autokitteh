package projects

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/psanford/memfs"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// returns (nil, nil) if no resources found.
func (p Projects) openProjectResourcesFS(ctx context.Context, projectID sdktypes.ProjectID) (fs.FS, error) {
	files, err := p.DB.GetProjectResources(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if files == nil {
		return nil, nil
	}

	memfs := memfs.New()

	for path, content := range files {
		if err := memfs.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return nil, fmt.Errorf("failed to create directory %q in memory filesystem: %w", filepath.Dir(path), err)
		}

		if err := memfs.WriteFile(path, content, 0o600); err != nil {
			return nil, fmt.Errorf("failed to write file %q to memory filesystem: %w", path, err)
		}
	}

	return memfs, nil
}
