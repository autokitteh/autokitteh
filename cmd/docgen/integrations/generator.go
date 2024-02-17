package integrations

import (
	"os"

	integrationsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/integrations/v1"
)

type Generator interface {
	Output() string
	Generate(akURL string, n int, i *integrationsv1.Integration)
}

func resetDir(dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	return os.MkdirAll(dir, 0o755) // rwxr-xr-x
}
