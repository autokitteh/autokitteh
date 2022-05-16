package main

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/urfave/cli/v2"

	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiprogram"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/lang"
)

func run(app *cli.App, args []string) int {
	if err := app.Run(args); err != nil {
		if e := (&apiprogram.Error{}); errors.As(err, &e) {
			fmt.Fprintf(os.Stderr, "program error: %v\n", e)
			return 10
		}

		if e := (&lang.ErrCanceled{}); errors.As(err, &e) {
			fmt.Fprintf(os.Stderr, "canceled: \n%v\n", apiprogram.SprintCallStack(e.CallStack))
			return 20
		}

		fmt.Fprintf(os.Stderr, "error: [%v] %v\n", reflect.TypeOf(errors.Unwrap(err)), err)
		return 1
	}

	return 0
}

func main() { os.Exit(run(newApp(), os.Args)) }
