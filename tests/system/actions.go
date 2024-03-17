package systest

import (
	"errors"
	"strings"
	"testing"
)

type akResult struct {
	output     string
	returnCode int
}

func runAction(t *testing.T, akPath, akAddr, step string) (any, error) {
	match := actions.FindStringSubmatch(step)
	switch match[1] {
	case "ak":
		args := append(ServiceUrlArg(akAddr), strings.Fields(match[3])...)
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
