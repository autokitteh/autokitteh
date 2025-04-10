package sdkruntimes

import (
	"context"
	"fmt"
	"io/fs"
	"net/url"
	"path/filepath"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkbuildfile"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func buildWithRuntime(
	ctx context.Context,
	rt sdkservices.Runtime,
	data *sdkbuildfile.RuntimeData,
	fs fs.FS,
	path string,
	symbols []sdktypes.Symbol,
) error {
	p, err := rt.Build(ctx, fs, path, symbols)
	if err != nil {
		return fmt.Errorf("build runtime: %w", err)
	}

	return data.MergeFrom(
		&sdkbuildfile.RuntimeData{
			Artifact: p,
		},
	)
}

func Build(
	ctx context.Context,
	rts sdkservices.Runtimes,
	srcFS fs.FS,
	symbols []sdktypes.Symbol,
	memo map[string]string,
) (*sdkbuildfile.BuildFile, error) {
	// requirements that are unsatisfiable at build time.
	externals := kittehs.Transform(symbols, func(sym sdktypes.Symbol) sdktypes.BuildRequirement {
		return sdktypes.NewBuildRequirement().WithSymbol(sym)
	})

	rtdescs, err := rts.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list runtimes: %w", err)
	}

	var q []sdktypes.BuildRequirement

	if err := fs.WalkDir(srcFS, ".", func(path string, de fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if de.IsDir() {
			return nil
		}

		if _, ok := MatchRuntimeByPath(rtdescs, path); !ok {
			// ignore non-runtime digestable files. if needed later on
			// during build, they will explicitly be added to the requirements.
			return nil
		}

		req := sdktypes.NewBuildRequirement().WithURL(&url.URL{Path: path})

		q = append(q, req)

		return nil
	}); err != nil {
		return nil, fmt.Errorf("walk dir: %w", err)
	}

	type rtCacheEntry struct {
		runtime sdkservices.Runtime
		data    *sdkbuildfile.RuntimeData
	}

	rtCache := make(map[string]*rtCacheEntry)
	visited := make(map[string]bool)

	for ; len(q) > 0; q = q[1:] {
		req := q[0]

		reqURL := req.URL()

		if reqURL.Scheme != "file" && reqURL.Scheme != "" {
			return nil, fmt.Errorf("unsupported scheme: %q", reqURL.Scheme)
		}

		path := reqURL.Path

		if visited[path] {
			continue
		}

		visited[path] = true

		rtd, ok := MatchRuntimeByPath(rtdescs, path)
		if !ok {
			externals = append(externals, req)
			continue
		}

		rtName := rtd.Name()

		var (
			rt   sdkservices.Runtime
			data *sdkbuildfile.RuntimeData
		)

		cached := rtCache[rtName.String()]
		if cached == nil {
			if rt, err = rts.New(ctx, rtName); err != nil {
				return nil, fmt.Errorf("new %q: %w", rtName, err)
			}

			data = &sdkbuildfile.RuntimeData{
				Info: sdkbuildfile.RuntimeInfo{
					Name: rt.Get().Name(),
				},
			}
		} else {
			rt = cached.runtime
			data = cached.data
		}

		if cached != nil {
			// we've already built using this runtime before.
			if !rtd.IsFilewiseBuild() {
				// ... we don't need to do it for each file.
				continue
			}
		} else if !rtd.IsFilewiseBuild() {
			fi, err := fs.Stat(srcFS, path)
			if err != nil {
				return nil, fmt.Errorf("stat %q: %w", path, err)
			}

			if !fi.IsDir() {
				// Use the dirname, not the filename.
				path = filepath.Dir(path)
			}
		}

		if err := buildWithRuntime(ctx, rt, data, srcFS, path, symbols); err != nil {
			return nil, fmt.Errorf("build error for %q: %w", path, err)
		}

		cached = &rtCacheEntry{runtime: rt, data: data}

		rtCache[rtName.String()] = cached

		q = append(q, cached.data.Artifact.Requirements()...)
	}

	rtData := kittehs.TransformMapToList(rtCache, func(_ string, c *rtCacheEntry) *sdkbuildfile.RuntimeData { return c.data })

	if externals == nil {
		externals = []sdktypes.BuildRequirement{}
	}

	return &sdkbuildfile.BuildFile{
		Info: sdkbuildfile.BuildInfo{
			Memo: memo,
		},
		Runtimes:            rtData,
		RuntimeRequirements: externals,
	}, nil
}
