package sdktypes

import (
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	runtimesv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runtimes/v1"
)

type ExportPB = runtimesv1.Export

type Export = *object[*ExportPB]

var (
	ExportFromProto       = makeFromProto(validateExport)
	StrictExportFromProto = makeFromProto(strictValidateExport)
	ToStrictExport        = makeWithValidator(strictValidateExport)
)

func strictValidateExport(pb *runtimesv1.Export) error {
	if err := ensureNotEmpty(pb.Symbol); err != nil {
		return err
	}

	return validateExport(pb)
}

func validateExport(pb *runtimesv1.Export) error {
	if _, err := ParseSymbol(pb.Symbol); err != nil {
		return fmt.Errorf("symbol: %w", err)
	}

	if _, err := CodeLocationFromProto(pb.Location); err != nil {
		return fmt.Errorf("location: %w", err)
	}

	return nil
}

func GetExportSymbol(e Export) Symbol { return kittehs.Must1(ParseSymbol(e.pb.Symbol)) }

func GetExportCodeLocation(e Export) CodeLocation {
	return kittehs.Must1(CodeLocationFromProto(e.pb.Location))
}

func NewExport(loc CodeLocation, sym Symbol) (Export, error) {
	return StrictExportFromProto(&ExportPB{Symbol: sym.String(), Location: loc.ToProto()})
}
