package manifest

import (
	"os"

	"gitlab.com/softkitteh/autokitteh/internal/pkg/akcue"
)

// TODO: validate manifest using cue.Value.Subsume
func ManifestFromPath(path string) (*Manifest, error) {
	var m Manifest

	if err := akcue.Load(path, &m); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}

	return &m, nil
}
