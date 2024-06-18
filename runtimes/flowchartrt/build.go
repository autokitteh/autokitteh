package flowchartrt

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"path/filepath"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/runtimes/flowchartrt/ast"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (rt) Build(ctx context.Context, fs fs.FS, path string, values []sdktypes.Symbol) (sdktypes.BuildArtifact, error) {
	data := make(map[string][]byte)

	type external struct {
		url *url.URL
		sym sdktypes.Symbol
	}

	var externals []*external

	visited := make(map[string]bool)

	var exports []sdktypes.BuildExport

	for q := []string{path}; len(q) != 0; q = q[1:] {
		path := q[0]

		if visited[path] {
			continue
		}

		visited[path] = true

		fr, err := fs.Open(path)
		if err != nil {
			return sdktypes.InvalidBuildArtifact, fmt.Errorf("open(%q): %w", path, err)
		}

		src, err := io.ReadAll(fr)

		_ = fr.Close()

		if err != nil {
			return sdktypes.InvalidBuildArtifact, fmt.Errorf("read(%q): %w", path, err)
		}

		f, err := ast.Parse(path, src)
		if err != nil {
			return sdktypes.InvalidBuildArtifact, err
		}

		if len(visited) == 1 {
			fexports, err := f.Exports()
			if err != nil {
				return sdktypes.InvalidBuildArtifact, fmt.Errorf("exports(%q): %w", path, err)
			}
			exports = append(exports, fexports...)
		}

		var buf bytes.Buffer
		if err := f.Write(&buf); err != nil {
			return sdktypes.InvalidBuildArtifact, fmt.Errorf("marshal(%q): %w", path, err)
		}

		data[path] = buf.Bytes()
		dir := filepath.Dir(path)

		for i, l := range f.Imports {
			lpath := l.Path

			if lpath == "" {
				return sdktypes.InvalidBuildArtifact, fmt.Errorf("empty load %d path", i)
			}

			if path != "" {
				lpath = filepath.Join(dir, lpath)
			}

			if isFlowchartPath(lpath) && lpath[0] != '@' {
				q = append(q, lpath)
				continue
			}

			externals = append(externals, &external{url: &url.URL{Path: lpath}})
		}
	}

	return sdktypes.BuildArtifactFromProto(
		&sdktypes.BuildArtifactPB{
			CompiledData: data,
			Exports:      kittehs.Transform(exports, sdktypes.ToProto),
			Requirements: kittehs.Transform(externals, func(x *external) *sdktypes.BuildRequirementPB {
				return &sdktypes.BuildRequirementPB{
					Url:    x.url.String(),
					Symbol: x.sym.String(),
				}
			}),
		},
	)
}
