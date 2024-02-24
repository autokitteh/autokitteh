package sdkbuildfile

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// Reads a build file from given reader.
// This does not make sure that all fields in BuildFile
// are present at the end of the read. Only that what is
// read is valid.
func Read(r io.Reader) (*BuildFile, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("gzip: %w", err)
	}

	defer gr.Close()

	tr := tar.NewReader(gr)

	var bf BuildFile

	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, fmt.Errorf("read next tar file: %w", err)
		}

		path := hdr.Name

		if !filepath.IsLocal(path) {
			return nil, fmt.Errorf("non-local path in tar %q", path)
		}

		buf := *bytes.NewBuffer(make([]byte, 0, hdr.Size))

		if _, err := io.Copy(&buf, tr); err != nil {
			return nil, fmt.Errorf("read %q: %w", path, err)
		}

		switch path {
		case filenames.version:
			if v := strings.TrimSpace(buf.String()); v != version {
				return nil, fmt.Errorf("unsupported build file format version %q != %q", v, version)
			}
		case filenames.info:
			if err := json.NewDecoder(&buf).Decode(&bf.Info); err != nil {
				return nil, fmt.Errorf("decode error for %q: %w", path, err)
			}
		case filenames.requirements:
			if err := json.NewDecoder(&buf).Decode(&bf.RuntimeRequirements); err != nil {
				return nil, fmt.Errorf("decode error for %q: %w", path, err)
			}
		default:
			// path structure: runtimes/[rtIndex]/[rest]
			// example: runtimes/name/compiled/examples/teststarlark/cats.star
			//   name = "name"
			//   rest = "compiled/examples/teststarlark/cats.star"
			parts := strings.SplitN(path, "/", 3)
			if len(parts) < 3 || parts[0] != filenames.runtimes {
				return nil, fmt.Errorf("unexpected path %q", path)
			}

			name, rest := parts[1], parts[2]

			n, err := sdktypes.ParseName(name)
			if err != nil {
				return nil, fmt.Errorf("invalid runtime name: %w", err)
			}

			if err := readRuntimeFile(&bf, n, rest, &buf); err != nil {
				return nil, fmt.Errorf("runtime %q: %w", name, err)
			}
		}
	}

	return &bf, nil
}

func readRuntimeFile(bf *BuildFile, name sdktypes.Name, path string, data *bytes.Buffer) error {
	rtIndex, rt := kittehs.FindFirst(bf.Runtimes, func(rt *RuntimeData) bool {
		return rt.Info.Name.String() == name.String()
	})

	if rt == nil {
		rt = &RuntimeData{Info: RuntimeInfo{Name: name}}
	}

	switch path {
	case filenames.info:
		if err := json.NewDecoder(data).Decode(&rt.Info); err != nil {
			return fmt.Errorf("decode error for %q: %w", path, err)
		}
	case filenames.resourcesIndex:
		// nop - nothing to do with it on read.
	case filenames.requirements:
		// TODO
	case filenames.exports:
		// TODO
	default:
		kind, rest, ok := strings.Cut(path, "/")
		if !ok {
			return fmt.Errorf("unexpected path %q", path)
		}

		switch kind {
		case filenames.compiledDir:
			if err := readRuntimeCompiledFile(rt, rest, data); err != nil {
				return fmt.Errorf("compiled %q: %w", path, err)
			}
		default:
			return fmt.Errorf("unexpected path %q", path)
		}
	}

	if rtIndex < 0 {
		bf.Runtimes = append(bf.Runtimes, rt)
	}

	return nil
}

func readRuntimeCompiledFile(rt *RuntimeData, path string, data *bytes.Buffer) (err error) {
	art := rt.Artifact
	if art == nil {
		art = kittehs.Must1(sdktypes.BuildArtifactFromProto(&sdktypes.BuildArtifactPB{}))
	}

	if art, err = art.Update(func(pb *sdktypes.BuildArtifactPB) {
		if pb.CompiledData == nil {
			pb.CompiledData = make(map[string][]byte)
		}

		pb.CompiledData[path] = data.Bytes()
	}); err != nil {
		return fmt.Errorf("invalid compiled file %q: %w", path, err)
	}

	rt.Artifact = art

	return
}
