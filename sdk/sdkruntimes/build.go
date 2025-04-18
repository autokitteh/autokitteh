package sdkruntimes

import (
	"context"
	"fmt"
	"io/fs"
	"net/url"

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

	filesPerRuntime := make(map[sdktypes.Symbol][]string) // rt name -> lexically sorted paths.

	if err := fs.WalkDir(srcFS, ".", func(path string, de fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("path %q: %w", path, err)
		}

		if de.IsDir() {
			return nil
		}

		rtName := MatchRuntimeByPath(rtdescs, path).Name()

		filesPerRuntime[rtName] = append(filesPerRuntime[rtName], path)

		return nil
	}); err != nil {
		return nil, fmt.Errorf("walk dir: %w", err)
	}

	var allData []*sdkbuildfile.RuntimeData

	for _, rtd := range rtdescs {
		rtName := rtd.Name()

		paths, ok := filesPerRuntime[rtd.Name()]
		if !ok {
			continue
		}

		data := &sdkbuildfile.RuntimeData{
			Info: sdkbuildfile.RuntimeInfo{
				Name: rtd.Name(),
			},
		}

		rt, err := rts.New(ctx, rtName)
		if err != nil {
			return nil, fmt.Errorf("new %q: %w", rtName, err)
		}

		if !rtd.IsFilewiseBuild() {
			if err := buildWithRuntime(ctx, rt, data, srcFS, ".", symbols); err != nil {
				return nil, fmt.Errorf("build error for %q: %w", ".", err)
			}
		} else {
			for _, path := range paths {
				if err := buildWithRuntime(ctx, rt, data, srcFS, path, symbols); err != nil {
					return nil, fmt.Errorf("build error for %q: %w", path, err)
				}
			}
		}

		allData = append(allData, data)
	}

	externals = append(externals, kittehs.Transform(filesPerRuntime[sdktypes.InvalidSymbol], func(path string) sdktypes.BuildRequirement {
		return sdktypes.NewBuildRequirement().WithURL(&url.URL{Path: path})
	})...)

	if externals == nil {
		externals = []sdktypes.BuildRequirement{}
	}

	return &sdkbuildfile.BuildFile{
		Info:                sdkbuildfile.BuildInfo{Memo: memo},
		Runtimes:            allData,
		RuntimeRequirements: externals,
	}, nil
}
