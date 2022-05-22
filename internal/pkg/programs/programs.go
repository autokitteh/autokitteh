package programs

import (
	"context"
	"fmt"

	"github.com/autokitteh/autokitteh/sdk/api/apiprogram"
	"github.com/autokitteh/autokitteh/sdk/api/apiproject"
	"github.com/autokitteh/autokitteh/internal/pkg/lang"
	"github.com/autokitteh/autokitteh/internal/pkg/lang/langtools"

	L "github.com/autokitteh/autokitteh/pkg/l"
)

type Programs struct {
	Catalog lang.Catalog

	// Only first that does not return nil is taken into account.
	// These apply only to paths without schemes. Versions are unaffected.
	PathRewriters []PathRewriterFunc

	// scheme -> loader
	CommonLoaders map[string]LoaderFunc

	L L.Nullable
}

func (p *Programs) SetCommonLoader(scheme string, loader LoaderFunc) {
	if p.CommonLoaders == nil {
		p.CommonLoaders = make(map[string]LoaderFunc)
	}

	p.CommonLoaders[scheme] = loader
}

func RewritePath(pathRewriters []PathRewriterFunc, path *apiprogram.Path) (*apiprogram.Path, error) {
	if path.Scheme() != "" {
		return path, nil
	}

	for _, f := range pathRewriters {
		if newPath, name, err := f(path.Path()); err != nil {
			return nil, fmt.Errorf("path rewriter %q: %w", name, err)
		} else if newPath != nil {
			if oldVersion := path.Version(); oldVersion != "" {
				if newVersion := newPath.Version(); newVersion != "" {
					// Cannot rewrite version.
					return nil, fmt.Errorf(
						"rewriter %q returned a version %q, but other version %q specified in input",
						name,
						oldVersion,
						newVersion,
					)
				}

				newPath = newPath.WithVersion(path.Version())
			}

			path = newPath
			break
		}
	}

	return path, nil
}

func (p *Programs) RewritePath(path *apiprogram.Path) (*apiprogram.Path, error) {
	return RewritePath(p.PathRewriters, path)
}

func (p *Programs) Load(
	ctx context.Context,
	pid apiproject.ProjectID, // use this to determine access?
	predecls []string,
	path *apiprogram.Path,
) (*apiprogram.Module, error) {
	l := p.L.With("path", path.String(), "project_id", pid)

	if len(p.CommonLoaders) == 0 {
		return nil, fmt.Errorf("no loaders configured")
	}

	actualPath, err := p.RewritePath(path)
	if err != nil {
		return nil, fmt.Errorf("path rewrite error: %w", err)
	}

	l = l.With("actual_path", actualPath)

	loader := p.CommonLoaders[actualPath.Scheme()]
	if loader == nil {
		return nil, fmt.Errorf("%q: no loader configured for scheme %q", actualPath, actualPath.Scheme())
	}

	src, err := loader(ctx, actualPath)
	if err != nil {
		return nil, fmt.Errorf("%q: %w", path, err)
	}

	// Use original path here to be compatible with loader in session (what the caller thinks
	// the path is).
	mod, _, err := langtools.CompileModule(ctx, p.Catalog, predecls, path, src)
	if err != nil {
		return nil, fmt.Errorf("%q: compile: %w", path, err)
	}

	l.Debug("module loaded", "lang", mod.Lang())

	return mod, nil
}
