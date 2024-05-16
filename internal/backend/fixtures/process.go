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
		"%s_%d__%d_%s__%s_%s",
		hostname,
		os.Getpid(),
		time.Now().Unix(),
		uuid.NewString(),
		version.Version,
		version.Commit,
	)

	processID = strings.ReplaceAll(processID, "-", "_")
	processID = strings.ReplaceAll(processID, ".", "_")
}
