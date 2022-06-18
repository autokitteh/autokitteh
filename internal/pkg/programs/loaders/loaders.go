package loaders

import (
	"context"
	"fmt"

	"go.autokitteh.dev/sdk/api/apiprogram"
	"go.autokitteh.dev/sdk/api/apiproject"

	"github.com/autokitteh/L"
)

type Loaders struct {

	// Only first that does not return nil is taken into account.
	// These apply only to paths without schemes. Versions are unaffected.
	PathRewriters []PathRewriterFunc

	// scheme -> loader
	CommonLoaders map[string]LoaderFunc

	L L.Nullable
}

func (p *Loaders) SetCommonLoader(scheme string, loader LoaderFunc) {
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

func (p *Loaders) RewritePath(path *apiprogram.Path) (*apiprogram.Path, error) {
	return RewritePath(p.PathRewriters, path)
}

func (p *Loaders) Fetch(
	ctx context.Context,
	pid apiproject.ProjectID, // use this to determine access?
	path *apiprogram.Path,
) ([]byte, string, error) {
	if path.IsRelative() {
		return nil, "", fmt.Errorf("relative paths are not supported by loader")
	}

	if len(p.CommonLoaders) == 0 {
		return nil, "", fmt.Errorf("no loaders configured")
	}

	actualPath, err := p.RewritePath(path)
	if err != nil {
		return nil, "", fmt.Errorf("path rewrite error: %w", err)
	}

	loader := p.CommonLoaders[actualPath.Scheme()]
	if loader == nil {
		return nil, "", fmt.Errorf("%q: no loader configured for scheme %q", actualPath, actualPath.Scheme())
	}

	src, ver, err := loader(ctx, actualPath)
	if err != nil {
		return nil, "", fmt.Errorf("%q: %w", path, err)
	}

	return src, ver, nil
}
