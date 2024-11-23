package systest

import (
	"os"
	"testing"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authusers"
)

var (
	consts = map[string]string{
		"DEFAULT_UID": authusers.DefaultUser.ID().String(),

		// Must correspond with `cmd/ak/common/exit.go`.
		"RC_NOT_FOUND":           "44",
		"RC_NOT_A_MEMBER":        "44",
		"RC_FAILED_PRECONDITION": "42",
		"RC_UNAUTHZ":             "43",
		"RC_UNAUTHN":             "41",
	}

	captures = make(map[string]string)
)

func captureJQ(t *testing.T, step string, ak *akResult, _ *httpResponse) error {
	match := jqCheck.FindStringSubmatch(step)
	name, query := match[1], match[2]

	v, err := jq(ak.output, query)
	if err != nil {
		return err
	}

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
		return consts[key]
	})
}
