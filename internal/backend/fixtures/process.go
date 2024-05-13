package fixtures

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"

	"go.autokitteh.dev/autokitteh/internal/version"
)

var processID string

func ProcessID() string { return processID }

func init() {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	processID = fmt.Sprintf(
		"%s_%d_%s_%s_%d_%s",
		strings.Map(func(r rune) rune {
			if r == '.' || r == '-' {
				return '_'
			}
			return r
		}, hostname),
		os.Getpid(),
		version.Version,
		version.Commit,
		time.Now().Unix(),
		strings.ReplaceAll(uuid.NewString(), "-", ""),
	)
}
