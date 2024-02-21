package common

import (
	"fmt"
	"io"
	"os"
)

func Consume(args []string) (data []byte, path string, err error) {
	switch len(args) {
	case 0:
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			err = fmt.Errorf("stdin: %w", err)
		}
	case 1:
		path = args[0]
		data, err = os.ReadFile(path)
	default:
		return nil, "", fmt.Errorf("too many arguments")
	}

	return
}
