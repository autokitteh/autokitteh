package remotert

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func createTar(fs fs.FS) ([]byte, error) {
	var buf bytes.Buffer
	w := tar.NewWriter(&buf)
	if err := w.AddFS(fs); err != nil {
		return nil, err
	}

	w.Close()
	return buf.Bytes(), nil
}

type Export struct {
	Name string
	File string
	Line int
}

func asBuildExport(e Export) sdktypes.BuildExport {
	pb := sdktypes.BuildExportPB{
		Symbol: e.Name,
		Location: &sdktypes.CodeLocationPB{
			Path: e.File,
			Row:  uint32(e.Line),
			Col:  1,
		},
	}

	b, _ := sdktypes.BuildExportFromProto(&pb)
	return b
}

// Returns sdktypes.ProgramErrorAsError if not internal error.
func (*svc) Build(ctx context.Context, fsys fs.FS, path string, symbols []sdktypes.Symbol) (sdktypes.BuildArtifact, error) {

	ffs, err := kittehs.NewFilterFS(fsys, func(entry fs.DirEntry) bool {
		return !strings.Contains(entry.Name(), "__pycache__")
	})
	if err != nil {
		return sdktypes.InvalidBuildArtifact, err
	}

	data, err := createTar(ffs)
	if err != nil {
		return sdktypes.InvalidBuildArtifact, err
	}

	export := Export{File: "main.py", Name: "on_http_get", Line: 4}
	sdkexport := asBuildExport(export)

	compiledData := map[string][]byte{
		"archive": data,
	}

	// UI requires file names in the compiled data.
	tf := tar.NewReader(bytes.NewReader(data))
	for {
		hdr, err := tf.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return sdktypes.InvalidBuildArtifact, err
		}

		if !strings.HasSuffix(hdr.Name, ".py") {
			continue
		}

		compiledData[hdr.Name] = nil
	}

	// buildExports := kittehs.Transform(, asBuildExport)
	var art sdktypes.BuildArtifact
	art = art.WithCompiledData(compiledData).WithExports([]sdktypes.BuildExport{sdkexport})

	// resp, err := runner.Start(ctx, &pb.StartRunnerRequest{BuildArtifact: data, Vars: nil, WorkerAddress: "host.docker.internal:9980"})
	// if err != nil {
	// 	return sdktypes.InvalidBuildArtifact, fmt.Errorf("staring runner %w", err)
	// }

	// if resp.Error != "" {
	// 	return sdktypes.InvalidBuildArtifact, fmt.Errorf("staring runner %s", resp.Error)
	// }

	// stopResp, err := runner.Stop(ctx, &pb.StopRequest{RunnerId: resp.RunnerId})
	// if err != nil {
	// 	fmt.Printf("failed stopping build runner: %s\n", err)
	// }

	// if stopResp.Error != "" {
	// 	fmt.Printf("failed stopping build runner: %s\n", err)
	// }

	return art, nil
}
