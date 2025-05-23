package systest

import (
	"encoding/csv"
	"errors"
	"strings"
	"testing"

	"go.autokitteh.dev/autokitteh/tests"
)

func splitToArgs(cmdArgs string) []string {
	cmdArgs = strings.TrimSpace(cmdArgs)
	r := csv.NewReader(strings.NewReader(cmdArgs))
	r.Comma = ' '       // space
	r.LazyQuotes = true // allow quotes to appear in string
	fields, _ := r.Read()
	return fields
}

func runAction(t *testing.T, akPath, akAddr string, i int, step string, cfg *testConfig) (any, error) {
	t.Logf("*** ACTION: line %d: %q", i+1, step)
	match := actions.FindStringSubmatch(step)
	switch match[1] {
	case "user":
		return nil, setUser(strings.TrimSpace(strings.TrimPrefix(step, match[1])))
	case "setenv":
		return nil, setEnv(strings.TrimSpace(strings.TrimPrefix(step, match[1])))
	case "ak":
		args := make([]string, len(cfg.AK.ExtraArgs))
		copy(args, cfg.AK.ExtraArgs)
		args = append(args, splitToArgs(match[3])...)
		return tests.RunAKClient(t, akPath, akAddr, token, 0, args)
	case "http get", "http post":
		method := strings.ToUpper(match[2])
		url, body, _ := strings.Cut(match[3], " ")
		return &httpRequest{method: method, url: url, body: body}, nil
	case "wait":
		return waitForSession(t, akPath, akAddr, step)
	default:
		return nil, errors.New("unhandled action")
	}
}
