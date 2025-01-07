package sdkbuildfile

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func writeJSON(tw *tar.Writer, name string, data any) error {
	return writeString(tw, name, string(kittehs.Must1(json.MarshalIndent(data, "", "  ")))+"\n")
}

func writeString(tw *tar.Writer, name string, data string) error {
	return writeBytes(tw, name, []byte(data))
}

func writeBytes(tw *tar.Writer, name string, data []byte) error {
	return write(tw, name, int64(len(data)), bytes.NewBuffer(data))
}

func write(tw *tar.Writer, name string, size int64, r io.Reader) error {
	if err := tw.WriteHeader(&tar.Header{
		Name: name,
		Mode: 0o600, // rw- --- ---
		Size: size,
	}); err != nil {
		return fmt.Errorf("%q: write_header: %w", name, err)
	}

	if _, err := io.CopyN(tw, r, size); err != nil {
		return fmt.Errorf("%q: write_data: %w", name, err)
	}

	return nil
}

func writeBuildData(tw *tar.Writer, root string, data map[string][]byte) error {
	ks := slices.Collect(maps.Keys(data))
	sort.Strings(ks)

	for _, k := range ks {
		if err := writeBytes(tw, filepath.Join(root, filenames.compiledDir, k), data[k]); err != nil {
			return err
		}
	}

	return nil
}

func writeRuntime(tw *tar.Writer, root string, rt *RuntimeData) error {
	if err := writeJSON(tw, filepath.Join(root, filenames.info), rt.Info); err != nil {
		return err
	}

	reqs := rt.Artifact.Requirements()
	if reqs == nil {
		reqs = []sdktypes.BuildRequirement{}
	}

	if err := writeJSON(tw, filepath.Join(root, filenames.requirements), reqs); err != nil {
		return err
	}

	exports := rt.Artifact.Exports()
	if exports == nil {
		exports = []sdktypes.BuildExport{}
	}

	if err := writeJSON(tw, filepath.Join(root, filenames.exports), exports); err != nil {
		return err
	}

	if err := writeBuildData(tw, root, rt.Artifact.CompiledData()); err != nil {
		return fmt.Errorf("compiled_data: %w", err)
	}

	return nil
}

func (bf *BuildFile) Write(w io.Writer) error {
	gw := gzip.NewWriter(w)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	if err := writeString(tw, filenames.version, versionPrefix+version+"\n"); err != nil {
		return err
	}

	if err := writeJSON(tw, filenames.info, bf.Info); err != nil {
		return err
	}

	reqs := bf.RuntimeRequirements
	if reqs == nil {
		reqs = []sdktypes.BuildRequirement{}
	}

	if err := writeJSON(tw, filenames.requirements, reqs); err != nil {
		return err
	}

	names := make(map[string]bool, len(bf.Runtimes))

	sort.Slice(bf.Runtimes, func(i, j int) bool {
		return strings.Compare(bf.Runtimes[i].Info.Name.String(), bf.Runtimes[j].Info.Name.String()) < 0
	})

	for _, rt := range bf.Runtimes {
		name := rt.Info.Name

		if names[name.String()] {
			return errors.New("multiple runtimes with the same name not allowed")
		}

		names[name.String()] = true

		if err := writeRuntime(tw, filepath.Join(filenames.runtimes, name.String()), rt); err != nil {
			return err
		}
	}

	return nil
}
