package systest

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func buildClient(t *testing.T) string {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	}()

	akRootDir := filepath.Dir(wd)
	if err := os.Chdir(akRootDir); err != nil {
		t.Fatalf("failed to switch to parent directory: %v", err)
	}

	output, err := exec.Command("make", "ak").CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build client: %v\n%s", err, output)
	}

	return filepath.Join(akRootDir, "bin", "ak")
}

func runClient(akPath string, args []string) (*akResult, error) {
	// Running in a subprocess, not a goroutine (like the
	// server), to ensure state isolation between executions.
	cmd := exec.Command(akPath, args...)
	output, err := cmd.CombinedOutput()

	r := &akResult{
		output:     strings.TrimSpace(string(output)),
		returnCode: cmd.ProcessState.ExitCode(),
	}

	ee := new(exec.ExitError)
	if errors.As(err, &ee) {
		err = nil
	}

	return r, err
}
