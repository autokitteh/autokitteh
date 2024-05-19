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
	// unfortunaly string.Fields will split everything by spaces, including quoted args. Let's use regexp instead
	// re := regexp.MustCompile(`[^\s"']+|([^\s"']*"([^"]*)"[^\s"']*)+|'([^']*)`). DO we need quotes in arg?
	re := regexp.MustCompile(`[^\s"']+|"[^"]*"|'[^']*'`)
	args := re.FindAllString(cmdArgs, -1)
	for i, arg := range args {
		arg = strings.Trim(arg, "\"'")
		args[i] = arg
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
