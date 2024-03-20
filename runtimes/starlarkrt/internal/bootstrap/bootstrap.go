package bootstrap

import (
	_ "embed"
	"strings"

	"go.starlark.net/starlark"
	"go.starlark.net/syntax"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const exportPrefix = "# EXPORT:"

var (
	//go:embed bootstrap.star
	source string

	Module  *starlark.Program
	Exports []sdktypes.Symbol // generated from all lines in source that begin with `exportPrefix`.
)

func init() {
	Exports = kittehs.Transform(kittehs.Filter(strings.Split(source, "\n"), func(s string) bool {
		return strings.HasPrefix(s, exportPrefix)
	}), func(s string) sdktypes.Symbol {
		return kittehs.Must1(sdktypes.StrictParseSymbol(strings.TrimSpace(s[len(exportPrefix):])))
	})

	isExport := kittehs.ContainedIn(kittehs.Transform(Exports, kittehs.ToString)...)

	_, Module = kittehs.Must2(starlark.SourceProgramOptions(
		&syntax.FileOptions{},
		"__bootstrap__",
		source,
		isExport,
	))
}

func Run(th *starlark.Thread, predecls starlark.StringDict) (starlark.StringDict, error) {
	return Module.Init(th, predecls)
}
