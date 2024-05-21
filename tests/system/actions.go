package systest

import (
	"errors"
	"regexp"
	"strings"
	"testing"

	"go.autokitteh.dev/autokitteh/config"
)

type akResult struct {
	output     string
	returnCode int
}

func splitToArgs(cmdArgs string) []string {
	re := regexp.MustCompile(`\w+=".*?"|\w+=[\w.:]+|"[^"]+"|[\w-.:/]+`)
	args := re.FindAllString(cmdArgs, -1)
	for i, arg := range args {
		if len(arg) >= 1 && arg[0] == '"' && arg[len(arg)-1] == '"' {
			args[i] = strings.Trim(arg, "\"")
		}
	}
	return args
}

func runAction(t *testing.T, akPath, akAddr, step string) (any, error) {
	match := actions.FindStringSubmatch(step)
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
	case "sleep":
		return nil, sleepFor(strings.TrimSpace(strings.TrimPrefix(step, match[1])))
	default:
		return nil, errors.New("unhandled action")
	}
}
