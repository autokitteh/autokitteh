package langcue

import (
	"fmt"
	"io/fs"

	"github.com/psanford/memfs"
	"go.dagger.io/dagger/compiler"

	cuemod "github.com/autokitteh/autokitteh/cue.mod"
)

func UnmarshalCue(src []byte, dst interface{}) error {
	srcfs := memfs.New()

	// TODO: accept configurable filename (for error messages?).
	srcfs.WriteFile("main.cue", src, 0644)

	// TODO: SECURITY: make sure this cannot access the fs.
	v, err := compiler.Build("/", map[string]fs.FS{
		// HACK: both keys should've been "/" or "", but obviously since this is a
		// map this can't happen. using "/" with "" allows to "add" them togethor.
		"/":       srcfs,
		"cue.mod": cuemod.FS,
	}, "main.cue")
	if err != nil {
		return err
	}

	if err := v.Decode(&dst); err != nil {
		return fmt.Errorf("cue value decode: %w", err)
	}

	return nil
}
