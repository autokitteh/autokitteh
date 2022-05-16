package parsecmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mattn/go-shellwords"
)

type Command struct {
	Name   string            `json:"name"`
	Args   []string          `json:"args"`
	Kwargs map[string]string `json:"kwargs"`
	Parts  []string          `json:"parts"`
}

// name a=1 b="2 3 4" c=3 d e "f \"g h\" i"
func Parse(text string) (*Command, error) {
	args, err := shellwords.NewParser().Parse(text)
	if err != nil {
		return nil, err
	}

	if len(args) == 0 {
		return nil, errors.New("must contain at least the command name")
	}

	cmd := Command{
		Name:   args[0],
		Args:   make([]string, 0, len(args[1:])),
		Kwargs: make(map[string]string, len(args[1:])),
		Parts:  args,
	}

	for _, arg := range args[1:] {
		parts := strings.SplitN(arg, "=", 2)

		if len(parts) == 1 {
			cmd.Args = append(cmd.Args, arg)
		} else {
			k, v := parts[0], parts[1]

			if vv, found := cmd.Kwargs[k]; found {
				cmd.Kwargs[k] = fmt.Sprintf("%s,%s", vv, v)
			} else {
				cmd.Kwargs[k] = v
			}
		}
	}

	return &cmd, nil
}
