package sdktypes

import (
	"errors"

	runtimesv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runtimes/v1"
)

type BuildExport struct {
	object[*BuildExportPB, BuildExportTraits]
}

var InvalidBuildExport BuildExport

type BuildExportPB = runtimesv1.Export

type BuildExportTraits struct{}

func (BuildExportTraits) Validate(m *BuildExportPB) error {
	return errors.Join(
		objectField[CodeLocation]("location", m.Location),
		symbolField("symbol", m.Symbol),
	)
}

func (BuildExportTraits) StrictValidate(m *BuildExportPB) error {
	return nonzeroMessage(m)
}

func BuildExportFromProto(m *BuildExportPB) (BuildExport, error) {
	return FromProto[BuildExport](m)
}

func NewBuildExport() BuildExport { return zeroObject[BuildExport]() }

func (r BuildExport) WithLocation(loc CodeLocation) BuildExport {
	return BuildExport{r.forceUpdate(func(pb *BuildExportPB) { pb.Location = loc.ToProto() })}
}

func (r BuildExport) WithSymbol(sym Symbol) BuildExport {
	return BuildExport{r.forceUpdate(func(pb *BuildExportPB) { pb.Symbol = sym.String() })}
}
