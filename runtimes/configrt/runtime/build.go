package runtime

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/proto"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/runtimes/configrt/parsers"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func Build(ctx context.Context, rootURL *url.URL, path string) (sdktypes.BuildArtifact, error) {
	if rootURL.Scheme != "file" && rootURL.Scheme != "" {
		return nil, fmt.Errorf("unsupported scheme: %q", rootURL.Scheme)
	}

	mainPath := rootURL.JoinPath(path).Path

	suffix := kittehs.MatchLongestSuffix(mainPath, parsers.ExtensionsWithDotPrefix)

	parser := parsers.Parsers[strings.TrimPrefix(suffix, ".")]
	if parser == nil {
		return nil, fmt.Errorf("unhandled extension for %q: %w", mainPath, sdkerrors.ErrInvalidArgument)
	}

	r, err := os.Open(mainPath)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", mainPath, err)
	}
	defer r.Close()

	compiled, err := parser(r)
	if err != nil {
		return nil, err
	}

	if !sdktypes.IsDictValue(compiled) && !sdktypes.IsStructValue(compiled) && !sdktypes.IsModuleValue(compiled) {
		return nil, fmt.Errorf("source represents niether a dict, struct or module")
	}

	exports, err := evaluateValue(compiled)
	if err != nil {
		return nil, fmt.Errorf("produced invalid data: %w", err)
	}

	data, err := proto.Marshal(sdktypes.ToMessage(compiled))
	if err != nil {
		return nil, fmt.Errorf("value marshal: %w", err)
	}

	return sdktypes.BuildArtifactFromProto(
		&sdktypes.BuildArtifactPB{
			Exports: kittehs.Transform(maps.Keys(exports), func(key string) *sdktypes.ExportPB {
				sym := kittehs.Must1(sdktypes.ParseSymbol(key))
				x := kittehs.Must1(sdktypes.NewExport(nil, sym))
				return x.ToProto()
			}),
			CompiledData: map[string][]byte{
				path: data,
			},
		},
	)
}
