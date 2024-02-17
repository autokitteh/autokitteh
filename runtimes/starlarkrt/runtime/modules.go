package runtime

import (
	"path/filepath"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

var (
	Extensions = []string{"star", "kitteh.star"}

	isExtension = kittehs.ContainedIn(Extensions...)
)

func isStarlarkPath(path string) bool {
	return isExtension(strings.TrimPrefix(filepath.Ext(path), "."))
}
