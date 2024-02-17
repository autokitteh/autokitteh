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
		args := []string{"--url=http://" + akAddr}
		args = append(args, strings.Fields(match[3])...)
		return runClient(akPath, args)
	case "http":
		return runActionHTTP(t, match[2], match[3])
	default:
		return nil, errors.New("unhandled action")
	}
}

// TODO: Return an actual HTTP response, not a dummy *string.
func runActionHTTP(t *testing.T, method, url string) (*string, error) {
	return nil, errors.New("not implemented yet")
}
