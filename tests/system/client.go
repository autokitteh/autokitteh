package systest

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"go.autokitteh.dev/autokitteh/config"
)

const (
	waitInterval = 100 * time.Millisecond
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

	akRootDir := filepath.Dir(filepath.Dir(wd))
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

func waitForSession(akPath, akAddr, step string) (string, error) {
	// Parse wait parameters.
	match := waitAction.FindStringSubmatch(step)
	if match == nil {
		return "", errors.New("invalid action")
	}
	duration, err := time.ParseDuration(match[1])
	if err != nil {
		return "", fmt.Errorf("invalid duration %q: %w", match[1], err)
	}
	id := match[2]

	// Check the session state with the AK client.
	state := regexp.MustCompile(`state:SESSION_STATE_TYPE_(COMPLETED|ERROR|STOPPED)`)
	args := append(config.ServiceUrlArg(akAddr), "session", "get", id)
	startTime := time.Now()

	for time.Since(startTime) < duration {
		result, err := runClient(akPath, args)
		if err != nil {
			return "", fmt.Errorf("failed to get session: %w", err)
		}
		if state.MatchString(result.output) {
			duration = time.Since(startTime).Round(time.Millisecond)
			return fmt.Sprintf("waited %s for session %s", duration, id), nil
		}
		time.Sleep(waitInterval)
	}

	text := fmt.Sprintf("session %s not done after %s", id, duration)

	args = append(config.ServiceUrlArg(akAddr), "event", "list", "--integration=http")
	result, err := runClient(akPath, args)
	if err == nil {
		text += fmt.Sprintf("\nEvent list:\n%s", result.output)
	}

	args = append(config.ServiceUrlArg(akAddr), "session", "list", "-J")
	result, err = runClient(akPath, args)
	if err == nil {
		text += fmt.Sprintf("\n---\nSession list:\n%s", result.output)
	}

	return "", errors.New(text)
}

func setEnv(args string) error {
	n, v, ok := strings.Cut(args, " ")
	if !ok {
		return errors.New("invalid setenv action")
	}

	n = strings.TrimSpace(n)
	v = strings.TrimSpace(v)

	if strings.HasPrefix(v, "\"") {
		var err error
		if v, err = strconv.Unquote(v); err != nil {
			return fmt.Errorf("failed to unquote value: %w", err)
		}
	}

	// TODO(ENG-666): Use t.Setenv() instead of os.Setenv().
	if err := os.Setenv(n, v); err != nil {
		return fmt.Errorf("failed to set environment variable: %w", err)
	}

	return nil
}
