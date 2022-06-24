package akcue

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	cueload "cuelang.org/go/cue/load"

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

	cfg := cueload.Config{Dir: "/"}

	if cfg.Overlay, err = overlays(cfg.Dir, overlay); err != nil {
		return fmt.Errorf("overlays: %w", err)
	}

	v, err := build(ctx, &cfg, actual)
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

// TODO: this should all be done in memory.
func Parse(ctx context.Context, src []byte, dst interface{}, tags []string) error {
	d, err := os.MkdirTemp("", "")
	if err != nil {
		return fmt.Errorf("mkdirtemp: %w", err)
	}

	defer os.RemoveAll(d)

	p := filepath.Join(d, "manifest.cue")

	if err := os.WriteFile(p, src, 0600); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return Load(ctx, p, dst, tags)
}

func Load(ctx context.Context, path string, dst interface{}, tags []string) error {
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

	cfg := cueload.Config{
		Dir:  actual,
		Tags: tags,
	}

	if cfg.Overlay, err = overlays(cfg.Dir, overlay); err != nil {
		return fmt.Errorf("overlays: %w", err)
	}

	v, err := build(ctx, &cfg, args...)

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
