package common

import (
	"errors"
	"fmt"
	"io"
	"os"
)

func Consume(args []string) (data []byte, err error) {
	switch len(args) {
	case 0:
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			err = fmt.Errorf("stdin: %w", err)
		}
	case 1:
		data, err = os.ReadFile(args[0])
		if err != nil {
			err = NewExitCodeError(NotFoundExitCode, err)
		}
	default:
		return nil, errors.New("too many arguments")
	}

	return
}
