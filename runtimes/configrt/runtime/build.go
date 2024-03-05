package runtime

import (
	"context"
	"fmt"
	"io/fs"
	"strings"

	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/proto"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/runtimes/configrt/parsers"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func Build(ctx context.Context, fs fs.FS, mainPath string) (sdktypes.BuildArtifact, error) {
	suffix := kittehs.MatchLongestSuffix(mainPath, parsers.ExtensionsWithDotPrefix)

	parser := parsers.Parsers[strings.TrimPrefix(suffix, ".")]
	if parser == nil {
		return sdktypes.InvalidBuildArtifact, sdkerrors.NewInvalidArgumentError("unhandled extension for %q", mainPath)
	}

	r, err := fs.Open(mainPath)
	if err != nil {
		return sdktypes.InvalidBuildArtifact, fmt.Errorf("read %q: %w", mainPath, err)
	}
	defer r.Close()

	compiled, err := parser(r)
	if err != nil {
		return sdktypes.InvalidBuildArtifact, err
	}

	if !compiled.IsDict() && !compiled.IsStruct() && !compiled.IsModule() {
		return sdktypes.InvalidBuildArtifact, fmt.Errorf("source represents niether a dict, struct or module")
	}

	exports, err := evaluateValue(compiled)
	if err != nil {
		return sdktypes.InvalidBuildArtifact, fmt.Errorf("produced invalid data: %w", err)
	}

	data, err := proto.Marshal(compiled.Message())
	if err != nil {
		return sdktypes.InvalidBuildArtifact, fmt.Errorf("value marshal: %w", err)
	}

	return sdktypes.BuildArtifactFromProto(
		&sdktypes.BuildArtifactPB{
			Exports: kittehs.Transform(maps.Keys(exports), func(key string) *sdktypes.BuildExportPB {
				sym := kittehs.Must1(sdktypes.ParseSymbol(key))
				return sdktypes.NewBuildExport().WithSymbol(sym).ToProto()
			}),
			CompiledData: map[string][]byte{
				mainPath: data,
			},
		},
	)
}
