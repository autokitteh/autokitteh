package akcue

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"go.dagger.io/dagger/compiler"

	cuemod "github.com/autokitteh/autokitteh/cue.mod"
)

// TODO: this doesn't work well, for some reason does not allow to import packages.
func LoadFS(ctx context.Context, lfs fs.FS, path string, dst interface{}) error {
	actual, member, _ := strings.Cut(path, ":")

	fi, err := fs.Stat(lfs, actual)
	if err != nil {
		return err
	}

	if !fi.IsDir() {
		actual = filepath.Dir(actual)
	}

	subfs, err := fs.Sub(lfs, actual)
	if err != nil {
		return err
	}

	actual = filepath.Join("/", actual)

	overlay := map[string]fs.FS{
		actual:                           subfs,
		filepath.Join(actual, "cue.mod"): cuemod.FS,
	}

	v, err := compiler.Build(ctx, "/", overlay, actual)
	if err != nil {
		return err
	}

	if member != "" {
		if v = v.Lookup(member); v == nil {
			return fmt.Errorf("%q not defined", member)
		}
	}

	if err := v.Decode(dst); err != nil {
		return fmt.Errorf("cue value decode: %w", err)
	}

	return nil
}

func Load(ctx context.Context, path string, dst interface{}) error {
	actual, member, _ := strings.Cut(path, ":")

	fi, err := os.Stat(actual)
	if err != nil {
		return err
	}

	var args []string

	if !fi.IsDir() {
		args = []string{filepath.Base(actual)}
		actual = filepath.Dir(actual)
	}

	overlay := map[string]fs.FS{
		path:                             os.DirFS(actual),
		filepath.Join(actual, "cue.mod"): cuemod.FS,
	}

	v, err := compiler.Build(ctx, actual, overlay, args...)
	if err != nil {
		return err
	}

	if member != "" {
		if v = v.Lookup(member); v == nil {
			return fmt.Errorf("%q not defined", member)
		}
	}

	if err := v.Decode(dst); err != nil {
		return fmt.Errorf("cue value decode: %w", err)
	}

	return nil
}
