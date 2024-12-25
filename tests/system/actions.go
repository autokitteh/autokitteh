package systest

import (
	"encoding/csv"
	"errors"
	"strings"
	"testing"
)

type akResult struct {
	output     string
	returnCode int
}

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
		args := append(serviceUrlArg(akAddr), cfg.AK.ExtraArgs...)
		args = append(args, splitToArgs(match[3])...)
		return runClient(akPath, args)
	case "http get", "http post":
		method := strings.ToUpper(match[2])
		url, body, _ := strings.Cut(match[3], " ")
		return &httpRequest{method: method, url: url, body: body}, nil
	case "wait":
		return waitForSession(akPath, akAddr, step)
	default:
		return nil, errors.New("unhandled action")
	}
}
