package systest

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authusers"
)

var (
	// Immutable constants that are available to the script.
	// Expanded prior to validation parsing at the beginning of each test.
	consts = map[string]string{
		"DEFAULT_UID": authusers.DefaultUser.ID().String(),

		// Must correspond with `cmd/ak/common/exit.go`.
		"RC_NOT_FOUND":           "44",
		"RC_NOT_A_MEMBER":        "44",
		"RC_FAILED_PRECONDITION": "42",
		"RC_UNAUTHZ":             "43",
		"RC_UNAUTHN":             "41",
		"RC_BAD_REQUEST":         "40",
	}

	// Text captured during test run.
	// Expanded during test run.
	captures = make(map[string]string)
)

func captureJQ(t *testing.T, step string, ak *akResult, _ *httpResponse) error {
	match := jqCheck.FindStringSubmatch(step)
	name, query := match[1], match[2]

	v, err := jq(ak.output, query)
	if err != nil {
		return fmt.Errorf("%w. input: %s", err, ak.output)
	}

	t.Logf("captured %q into %q", v, name)

	captures[name] = v

	return nil
}

func captureRE(t *testing.T, step string, ak *akResult, _ *httpResponse) error {
	match := reCheck.FindStringSubmatch(step)
	name, query := match[1], match[2]

	re, err := regexp.Compile(query)
	if err != nil {
		return fmt.Errorf("failed to compile regexp %q: %w", query, err)
	}

	v := re.FindString(ak.output)

	t.Logf("captured %q into %q", v, name)

	captures[name] = v

	return nil
}

func expandCapture(s string) string {
	return os.Expand(s, func(key string) string {
		return captures[key]
	})
}

func expandConsts(s string) string {
	return os.Expand(s, func(key string) string {
		if v, ok := consts[key]; ok {
			return v
		}

		// make it to expandCapture.
		return "$" + key
	})
}
