package systest

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"go.autokitteh.dev/autokitteh/tests"
)

const (
	waitInterval = 100 * time.Millisecond
)

var (
	sessionStateFinal = regexp.MustCompile(`state:SESSION_STATE_TYPE_(COMPLETED|ERROR|STOPPED)`)
	sessionStateAll   = regexp.MustCompile(`state:SESSION_STATE_TYPE_`)
)

func waitForSession(t *testing.T, akPath, akAddr, step string) (string, error) {
	// Parse wait parameters.
	match := waitAction.FindStringSubmatch(step)
	if match == nil {
		return "", errors.New("invalid action")
	}
	duration, err := time.ParseDuration(match[1])
	if err != nil {
		return "", fmt.Errorf("invalid duration %q: %w", match[1], err)
	}

	waitType := match[2]
	id := match[3]

	stateRegex := sessionStateAll // wait .. unless .. session, wait for any session state
	isSessionExpected := waitType == "for"
	if isSessionExpected {
		stateRegex = sessionStateFinal // wait .. for .. session
	}

	// Check the session state with the AK client.
	args := []string{"session", "get", id}
	startTime := time.Now()

	sessionFound := false
	for time.Since(startTime) < duration {
		result, err := tests.RunAKClient(t, akPath, akAddr, token, 0, args)
		if err != nil {
			return "", fmt.Errorf("failed to get session: %w", err)
		}
		if sessionFound = stateRegex.MatchString(result.Output); sessionFound {
			duration = time.Since(startTime).Round(time.Millisecond)
			break
		}
		time.Sleep(waitInterval)
	}

	if isSessionExpected == sessionFound {
		return fmt.Sprintf("waited %s %s session %s. Session was found: %t", duration, waitType, id, sessionFound), nil
	}

	// Error handling.
	text := fmt.Sprintf("session %s not done after %s", id, duration)

	args = []string{"event", "list", "--integration=http"}
	result, err := tests.RunAKClient(t, akPath, akAddr, token, 0, args)
	if err == nil {
		text += "\nEvent list:\n" + result.Output
	}

	args = []string{"session", "list", "-J"}
	result, err = tests.RunAKClient(t, akPath, akAddr, token, 0, args)
	if err == nil {
		text += "\n---\nSession list:\n" + result.Output
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
