package systest

import (
	"encoding/csv"
	"errors"
	"strings"
	"testing"

	"go.autokitteh.dev/autokitteh/config"
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

func runAction(t *testing.T, akPath, akAddr, step string) (any, error) {
	match := actions.FindStringSubmatch(step)
	t.Log("action:", step)
	switch match[1] {
	case "setenv":
		return nil, setEnv(strings.TrimSpace(strings.TrimPrefix(step, match[1])))
	case "ak":
		args := append(config.ServiceUrlArg(akAddr), splitToArgs(match[3])...)
		return runClient(akPath, args)
	case "http get", "http post":
		method := strings.ToUpper(match[2])
		return &httpRequest{method: method, url: match[3]}, nil
	case "wait":
		return waitForSession(akPath, akAddr, step)
	default:
		return nil, errors.New("unhandled action")
	}
}
