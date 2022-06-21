package manifest

import (
	"context"
	"os"

	"github.com/autokitteh/autokitteh/internal/pkg/akcue"
)

// TODO: validate manifest using cue.Value.Subsume
func ManifestFromPath(ctx context.Context, path string) (*Manifest, error) {
	var m Manifest

	if err := akcue.Load(ctx, path, &m); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}

	return &m, nil
}

func ParseManifest(ctx context.Context, src []byte) (*Manifest, error) {
	var m Manifest

	if err := akcue.Parse(ctx, src, &m); err != nil {
		return nil, err
	}

	return &m, nil
}
