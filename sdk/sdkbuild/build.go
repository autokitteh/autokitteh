package sdkbuild

import (
	"context"
	"fmt"
	"io/fs"
	"net/url"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkbuild/sdkbuildfile"
	"go.autokitteh.dev/autokitteh/sdk/sdkruntimes"
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
		return fmt.Errorf("runtime_build: %w", err)
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
	externals, err := kittehs.TransformError(symbols, func(sym sdktypes.Symbol) (sdktypes.Requirement, error) {
		return sdktypes.NewRequirement(nil, nil, sym)
	})
	if err != nil {
		return nil, fmt.Errorf("invalid symbols: %w", err)
	}

	rtdescs, err := rts.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list: %w", err)
	}

	var q []sdktypes.Requirement

	if err := fs.WalkDir(srcFS, ".", func(path string, de fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if de.IsDir() {
			return nil
		}

		rtd := sdkruntimes.MatchRuntimeByPath(rtdescs, path)
		if rtd == nil {
			// ignore non-runtime digestable files. if needed later on
			// during build, they will explicitly be added to the requirements.
			return nil
		}

		req, err := sdktypes.NewRequirement(nil, &url.URL{Path: path}, nil)
		if err != nil {
			return err
		}

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

		reqURL := sdktypes.GetRequirementURL(req)

		if reqURL.Scheme != "file" && reqURL.Scheme != "" {
			return nil, fmt.Errorf("unsupported scheme: %q", reqURL.Scheme)
		}

		path := reqURL.Path

		if visited[path] {
			continue
		}

		visited[path] = true

		rtd := sdkruntimes.MatchRuntimeByPath(rtdescs, path)

		if rtd == nil {
			externals = append(externals, req)
			continue
		}

		rtName := sdktypes.GetRuntimeName(rtd)

		cached := rtCache[rtName.String()]
		if cached == nil {
			rt, err := rts.New(ctx, rtName)
			if err != nil {
				return nil, fmt.Errorf("new %q: %w", rtName, err)
			}

			cached = &rtCacheEntry{
				runtime: rt,
				data: &sdkbuildfile.RuntimeData{
					Info: sdkbuildfile.RuntimeInfo{
						Name: sdktypes.GetRuntimeName(rt.Get()),
					},
				},
			}
		}

		if err := buildWithRuntime(ctx, cached.runtime, cached.data, srcFS, path, symbols); err != nil {
			return nil, fmt.Errorf("build error for %q: %w", path, err)
		}

		rtCache[rtName.String()] = cached

		q = append(q, sdktypes.GetBuildArtifactRequirements(cached.data.Artifact)...)
	}

	rtDatas := kittehs.TransformMapToList(rtCache, func(_ string, c *rtCacheEntry) *sdkbuildfile.RuntimeData { return c.data })

	if externals == nil {
		externals = []sdktypes.Requirement{}
	}

	return &sdkbuildfile.BuildFile{
		Info: sdkbuildfile.BuildInfo{
			Memo: memo,
		},
		Runtimes:            rtDatas,
		RuntimeRequirements: externals,
	}, nil
}
