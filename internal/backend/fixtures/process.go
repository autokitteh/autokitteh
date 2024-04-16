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
		"%s_%d_%s_%d_%s",
		hostname,
		os.Getpid(),
		version.Commit,
		time.Now().Unix(),
		strings.ReplaceAll(uuid.NewString(), "-", ""),
	)
}
