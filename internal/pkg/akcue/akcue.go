package akcue

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"go.dagger.io/dagger/compiler"

	cuemod "gitlab.com/softkitteh/autokitteh/cue.mod"
)

func Load(path string, dst interface{}) error {
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

	v, err := compiler.Build(actual, overlay, args...)
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
