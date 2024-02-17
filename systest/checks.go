package systest

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func runCheck(step string, ak *akResult, httpResp *string) error {
	match := steps.FindStringSubmatch(step)
	switch match[1] {
	case "output":
		return checkAKOutput(step, ak)
	case "return":
		return checkAKReturnCode(step, ak)
	default:
		return errors.New("unhandled check")
	}
}

func checkAKOutput(step string, ak *akResult) error {
	match := akCheckOutput.FindStringSubmatch(step)
	want := strings.TrimSpace(match[3])
	got := ak.output

	if strings.HasPrefix(match[2], "file") {
		b, err := os.ReadFile(want)
		if err != nil {
			return fmt.Errorf("failed to read embedded file: %w", err)
		}
		want = strings.TrimSpace(string(b))
	}

	switch match[1] {
	case "equals":
		if want != got {
			return stringCheckFailed(want, got)
		}
	case "contains":
		if !strings.Contains(got, want) {
			return stringCheckFailed(want, got)
		}
	case "regex":
		matched, err := regexp.MatchString(want, got)
		if err != nil {
			return fmt.Errorf("failed to match regex: %w", err)
		}
		if !matched {
			return stringCheckFailed(want, got)
		}
	default:
		return errors.New("unhandled AK check type")
	}
	return nil
}

func checkAKReturnCode(step string, ak *akResult) error {
	match := akCheckReturn.FindStringSubmatch(step)
	expected, err := strconv.Atoi(match[1])
	if err != nil {
		return fmt.Errorf("failed to parse expected return code: %w", err)
	}
	if expected != ak.returnCode {
		var sb strings.Builder
		// Test log will already show what was expected, where, and why.
		sb.WriteString(fmt.Sprintf("got return code %d", ak.returnCode))
		// Append the AK output for context, if there is any.
		if ak.output != "" {
			sb.WriteString("\n" + ak.output)
		}
		return fmt.Errorf(sb.String())
	}
	return nil
}

// We implement our own string checks and return errors on failures
// instead of "github.com/stretchr/testify/assert" because:
// 1. It results in shorter, simpler, more readable error messages
// 2. Fail-fast behavior (no point in subsequent actions and checks)
func stringCheckFailed(want, got string) error {
	var sb strings.Builder

	sb.WriteString("\n--- Expected: ")
	if strings.Contains(want, "\n") {
		sb.WriteString("\n    ")
	}
	sb.WriteString(strings.ReplaceAll(want, "\n", "\n    "))

	sb.WriteString("\n+++ Actual:   ")
	if strings.Contains(got, "\n") {
		sb.WriteString("\n    ")
	}
	sb.WriteString(strings.ReplaceAll(got, "\n", "\n    "))

	return fmt.Errorf(sb.String())
}
