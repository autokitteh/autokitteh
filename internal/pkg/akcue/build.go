// Based on https://github.com/dagger/dagger/blob/main/compiler/build.go.
package akcue

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"path/filepath"

	cueload "cuelang.org/go/cue/load"

	"go.dagger.io/dagger/compiler"
)

func overlays(dir string, ovlys map[string]fs.FS) (map[string]cueload.Source, error) {
	ret := make(map[string]cueload.Source)

	// Map the source files into the overlay
	for mnt, f := range ovlys {
		f := f
		mnt := mnt
		err := fs.WalkDir(f, ".", func(p string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !entry.Type().IsRegular() {
				return nil
			}

			if filepath.Ext(entry.Name()) != ".cue" {
				return nil
			}

			contents, err := fs.ReadFile(f, p)
			if err != nil {
				return fmt.Errorf("%s: %w", p, err)
			}

			overlayPath := path.Join(dir, mnt, p)
			ret[overlayPath] = cueload.FromBytes(contents)
			return nil
		})
		if err != nil {
			return nil, compiler.Err(err)
		}
	}

	return ret, nil
}

func build(
	ctx context.Context,
	cfg *cueload.Config,
	args ...string,
) (*compiler.Value, error) {
	instances := cueload.Instances(args, cfg)
	if len(instances) != 1 {
		return nil, errors.New("only one package is supported at a time")
	}

	instance := instances[0]
	if err := instance.Err; err != nil {
		return nil, compiler.Err(err)
	}

	c := compiler.DefaultCompiler

	v := c.Context.BuildInstance(instance)
	if err := v.Err(); err != nil {
		return nil, c.Err(err)
	}

	if err := v.Validate(); err != nil {
		return nil, c.Err(err)
	}

	return compiler.Wrap(v), nil
}
