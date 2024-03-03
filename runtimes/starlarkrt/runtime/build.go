package runtime

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"path/filepath"

	"go.starlark.net/starlark"
	"go.starlark.net/syntax"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/runtimes/starlarkrt/internal/bootstrap"
	"go.autokitteh.dev/autokitteh/runtimes/starlarkrt/internal/libs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var fileOptions = &syntax.FileOptions{
	Set:       true,
	While:     true,
	Recursion: true,
}

func makeExport(path, name string, pos syntax.Position) sdktypes.BuildExport {
	loc := sdktypes.CodeLocationPB{
		Path: path,
	}

	if pos.IsValid() {
		loc.Row = uint32(pos.Line)
		loc.Col = uint32(pos.Col)
	}

	exp, err := sdktypes.BuildExportFromProto(&sdktypes.BuildExportPB{
		Location: &loc,
		Symbol:   name,
	})
	if err != nil {
		return sdktypes.InvalidBuildExport
	}

	return exp
}

func getExports(path string, f *syntax.File) (exports []sdktypes.BuildExport) {
	for _, stmt := range f.Stmts {
		switch stmt := stmt.(type) {
		case *syntax.DefStmt:
			if stmt.Name != nil {
				pos, _ := stmt.Span()
				exports = append(exports, makeExport(path, stmt.Name.Name, pos))
			}
		case *syntax.AssignStmt:
			switch lhs := stmt.LHS.(type) {
			case *syntax.Ident:
				pos, _ := lhs.Span()
				exports = append(exports, makeExport(path, lhs.Name, pos))
			case *syntax.TupleExpr, *syntax.ListExpr:
				// TODO(ENG-199)
			}
		}
	}

	return
}

// TODO: we might want to stream the build product data as it might be big? Or we just limit the build size.
func Build(ctx context.Context, fs fs.FS, path string, symbols []sdktypes.Symbol) (sdktypes.BuildArtifact, error) {
	data := make(map[string][]byte)

	type external struct {
		url *url.URL
		sym sdktypes.Symbol
		pos *syntax.Position
	}

	externals := kittehs.Transform(symbols, func(sym sdktypes.Symbol) *external {
		return &external{sym: sym}
	})

	// build list of preloaded/global symbols
	symbols = append(symbols, bootstrap.Exports...) // bootstrap symbols

	symImplMap := libs.LoadModules(0) // seed must be a constant here.
	for sym := range symImplMap {
		symbols = append(symbols, kittehs.Must1(sdktypes.ParseSymbol(sym)))
	}

	isPredecl := kittehs.ContainedIn(kittehs.Transform(symbols, kittehs.ToString[sdktypes.Symbol])...)

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

		f, mod, err := starlark.SourceProgramOptions(fileOptions, path, src, isPredecl)
		if err != nil {
			return sdktypes.InvalidBuildArtifact, translateError(err, map[string]string{
				"source_path": path,
			})
		}

		if len(visited) == 1 {
			exports = append(exports, getExports(path, f)...)
		}

		var modBuf bytes.Buffer
		if err := mod.Write(&modBuf); err != nil {
			return sdktypes.InvalidBuildArtifact, fmt.Errorf("mod.write: %w", err)
		}

		data[path] = modBuf.Bytes()

		for i := 0; i < mod.NumLoads(); i++ {
			lpath, pos := mod.Load(i)

			if path != "" {
				lpath = filepath.Join(filepath.Dir(path), lpath)
			}

			if isStarlarkPath(lpath) {
				q = append(q, lpath)
				continue
			}

			externals = append(externals, &external{url: &url.URL{Path: lpath}, pos: &pos})
		}
	}

	return sdktypes.BuildArtifactFromProto(
		&sdktypes.BuildArtifactPB{
			CompiledData: data,
			Exports:      kittehs.Transform(exports, sdktypes.ToProto),
			Requirements: kittehs.Transform(externals, func(x *external) *sdktypes.BuildRequirementPB {
				var loc *sdktypes.CodeLocationPB

				if x.pos != nil {
					loc = &sdktypes.CodeLocationPB{
						Path: x.pos.Filename(),
						Row:  uint32(x.pos.Line),
						Col:  uint32(x.pos.Col),
					}
				}

				return &sdktypes.BuildRequirementPB{
					Url:      x.url.String(),
					Symbol:   x.sym.String(),
					Location: loc,
				}
			}),
		},
	)
}
