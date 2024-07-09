package ast

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func init() {
	gob.Register([]any{})
	gob.Register(map[string]any{})
}

func decode(path string, src []byte) (*Flowchart, error) {
	var (
		f   Flowchart
		err error
	)

	switch filepath.Ext(path) {
	case ".json":
		d := json.NewDecoder(bytes.NewBuffer(src))
		d.DisallowUnknownFields()
		err = d.Decode(&f)
	case ".yaml", ".yml":
		d := yaml.NewDecoder(bytes.NewBuffer(src))
		d.KnownFields(true)
		err = d.Decode(&f)
	default:
		err = fmt.Errorf("unsupported file extension: %q", filepath.Ext(path))
	}

	if err != nil {
		return nil, err
	}

	return &f, nil
}

func Parse(path string, src []byte) (*Flowchart, error) {
	f, err := decode(path, src)
	if err != nil {
		return nil, err
	}

	if err := f.preprocess(path); err != nil {
		return nil, err
	}

	if err := f.Validate(); err != nil {
		// TODO: program error.
		return nil, err
	}

	// TODO: Validate references.

	return f, nil
}

func (f *Flowchart) Write(w io.Writer) error { return gob.NewEncoder(w).Encode(f) }

func Read(path string, r io.Reader) (*Flowchart, error) {
	var f Flowchart

	if err := gob.NewDecoder(r).Decode(&f); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	if err := f.preprocess(path); err != nil {
		return nil, err
	}

	if err := f.Validate(); err != nil {
		return nil, err
	}

	return &f, nil
}

func (f *Flowchart) preprocess(path string) error {
	f.path = path

	for _, n := range f.Nodes {
		n.loc, _ = sdktypes.NewCodeLocation(n.Name, f.path, 0, 0)
	}

	return nil
}
