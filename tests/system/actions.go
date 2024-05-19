package systest

import (
	"errors"
	"strings"
	"testing"

	"go.autokitteh.dev/autokitteh/config"
)

type akResult struct {
	output     string
	returnCode int
}

func runAction(t *testing.T, akPath, akAddr, step string) (any, error) {
	match := actions.FindStringSubmatch(step)
	switch match[1] {
	case "setenv":
		return nil, setEnv(strings.TrimSpace(strings.TrimPrefix(step, match[1])))
	case "ak":
		args := append(config.ServiceUrlArg(akAddr), strings.Fields(match[3])...)
		return runClient(akPath, args)
	case "http get", "http post":
		method := strings.ToUpper(match[2])
		return &httpRequest{method: method, url: match[3]}, nil
	case "wait":
		return waitForSession(akPath, akAddr, step)
	case "sleep":
		return nil, sleepFor(strings.TrimSpace(strings.TrimPrefix(step, match[1])))
	default:
		return nil, errors.New("unhandled action")
	}
}
