package projects

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"

	"github.com/psanford/memfs"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// If srcURL is nil, get resources from the database.
// If srcURL scheme is empty, get resources from the database. If none found, get from filesystem, if allowed.
// If srcURL scheme is "file", get resources from filesystem, if allowed.
// If srcURL scheme is "project", get resources from database.
func (p Projects) openResourcesFS(ctx context.Context, projectID sdktypes.ProjectID, srcURL *url.URL) (fs.FS, error) {
	if srcURL == nil || *srcURL == (url.URL{}) {
		return p.openResourcesFS(ctx, projectID, &url.URL{Scheme: "project"})
	}

	switch srcURL.Scheme {
	case "":
		fs, err := p.openProjectResourcesFS(ctx, projectID, srcURL.Path)
		if err != nil {
			return nil, err
		}
		if fs != nil || !p.Config.Resources.AllowLocalFS {
			return fs, nil
		}
		fallthrough
	case "file":
		if !p.Config.Resources.AllowLocalFS {
			return nil, errors.New("resources from local filesystem are not allowed")
		}
		return os.DirFS(srcURL.Path), nil
	case "project":
		return p.openProjectResourcesFS(ctx, projectID, srcURL.Path)
	default:
		return nil, fmt.Errorf("%w: unhandled resources url scheme %q", sdkerrors.ErrNotImplemented, srcURL.Scheme)
	}
}

// returns nil, nil if no resources found.
func (p Projects) openProjectResourcesFS(ctx context.Context, projectID sdktypes.ProjectID, path string) (fs.FS, error) {
	files, err := p.DB.GetProjectResources(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if files == nil {
		return nil, nil
	}

	memfs := memfs.New()

	for path, content := range files {
		if err := memfs.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, fmt.Errorf("failed to create directory %q in memory filesystem: %w", filepath.Dir(path), err)
		}

		if err := memfs.WriteFile(path, content, 0o644); err != nil {
			return nil, fmt.Errorf("failed to write file %q to memory filesystem: %w", path, err)
		}
	}

	return memfs.Sub(path)
}
