package programs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.autokitteh.dev/sdk/api/apiplugin"
	"go.autokitteh.dev/sdk/api/apiprogram"
	"go.autokitteh.dev/sdk/api/apiproject"

	"github.com/autokitteh/L"

	"github.com/autokitteh/autokitteh/internal/pkg/lang"
	"github.com/autokitteh/autokitteh/internal/pkg/lang/langtools"
	"github.com/autokitteh/autokitteh/internal/pkg/programs/loaders"
	"github.com/autokitteh/autokitteh/internal/pkg/programsstore"
)

var bootPath = apiprogram.MustParsePathString("$internal:boot.kitteh")

type File = programsstore.File

type Programs struct {
	Store   *programsstore.Store
	Loaders *loaders.Loaders
	Catalog lang.Catalog
	L       L.Nullable
}

func (p *Programs) Update(
	ctx context.Context,
	pid apiproject.ProjectID,
	files []*File,
) error {
	return p.Store.Update(ctx, pid, files)
}

func (p *Programs) Get(ctx context.Context, pid apiproject.ProjectID, path *apiprogram.Path) (*File, error) {
	fs, err := p.Store.Get(ctx, pid, []*apiprogram.Path{path}, false)
	if err != nil {
		return nil, err
	}

	if len(fs) == 0 || errors.Is(err, programsstore.ErrNotFound) {
		return nil, nil
	}

	if len(fs) > 1 {
		return nil, fmt.Errorf("%d records returned", len(fs))
	}

	return fs[0], nil
}

type FetchResult struct {
	Files     []*File
	PluginIDs []apiplugin.PluginID
}

func (fr *FetchResult) Paths() (paths []*apiprogram.Path) {
	for _, f := range fr.Files {
		paths = append(paths, f.Path.WithVersion(f.FetchedVersion))
	}
	return
}

func (fr *FetchResult) Modules() (mods []*apiprogram.Module) {
	for _, f := range fr.Files {
		mods = append(mods, f.Module)
	}
	return
}

func (p *Programs) Fetch(
	ctx context.Context,
	pid apiproject.ProjectID,
	mainPath *apiprogram.Path,
	predecls []string,
) (*FetchResult, error) {
	q := []*apiprogram.Path{bootPath}

	var (
		batch     []*File
		pluginIDs []apiplugin.PluginID
	)

	for ; len(q) != 0; q = q[1:] {
		path := q[0]

		// TODO: [# internal_main #] duplicity
		if path.String() == "$internal:main" {
			path = mainPath
		}

		src, ver, err := p.Loaders.Fetch(ctx, pid, path)
		if err != nil {
			return nil, fmt.Errorf("fetch %v %q: %w", pid, path, err)
		}

		mod, _, err := langtools.CompileModule(ctx, p.Catalog, predecls, path, src)
		if err != nil {
			return nil, fmt.Errorf("compile %v %q: %w", pid, path, err)
		}

		batch = append(batch, &File{
			Path:           path,
			FetchedVersion: ver,
			Source:         src,
			Module:         mod,
			FetchedAt:      time.Now(),
		})

		deps, err := langtools.GetModuleDependencies(ctx, p.Catalog, mod)
		if err != nil {
			return nil, fmt.Errorf("get deps %v %q: %w", pid, path, err)
		}

		for _, dep := range deps {
			if dep.IsInternal() {
				if !path.IsInternal() {
					return nil, fmt.Errorf("internal loads are allowed only from internal modules")
				}
			} else if plugID, isPlugin := dep.PluginID(); isPlugin {
				// handled by either the lang load or session load (plugin).
				pluginIDs = append(pluginIDs, plugID)
				continue
			}

			dep, err := apiprogram.JoinWithParent(path, dep)
			if err != nil {
				return nil, fmt.Errorf("invalid relative path %q to %q: %w", dep.String(), path.String(), err)
			}

			q = append(q, dep)
		}
	}

	if err := p.Update(ctx, pid, batch); err != nil {
		p.L.Error("update failed", "err", err)
	}

	return &FetchResult{Files: batch, PluginIDs: pluginIDs}, nil
}
