package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/qri-io/starlib"
	"go.starlark.net/repl"
	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarktest"
)

var (
	showenv  = flag.Bool("showenv", false, "on success, print final global environment")
	execprog = flag.String("c", "", "execute program `prog`")
)

func main() { os.Exit(doMain()) }

type reporter struct{ count int }

func (r *reporter) Error(args ...interface{}) {
	r.count++
	log.Print(args...)
}

func doMain() int {
	resolve.AllowGlobalReassign = true

	log.SetPrefix("starsh: ")
	log.SetFlags(0)

	predecls := make(map[string]starlark.Value)
	flag.Func("p", "", func(v string) error {
		parts := strings.SplitN(v, "=", 2)

		if len(parts) < 2 {
			return fmt.Errorf("predecls must be in k=v format")
		}

		predecls[parts[0]] = starlark.String(parts[1])

		return nil
	})

	flag.Parse()

	replLoader := repl.MakeLoad()

	reporter := &reporter{}

	thread := &starlark.Thread{
		Load: func(thread *starlark.Thread, module string) (starlark.StringDict, error) {
			if module == "assert" {
				starlarktest.SetReporter(thread, reporter)
				return starlarktest.LoadAssertModule()
			}

			if m, err := starlib.Loader(thread, module); err == nil {
				return m, nil
			}

			return replLoader(thread, module)
		},
	}
	globals := make(starlark.StringDict)

	initModules()

	switch {
	case flag.NArg() == 1 || *execprog != "":
		var (
			filename string
			src      interface{}
			err      error
		)
		if *execprog != "" {
			// Execute provided program.
			filename = "cmdline"
			src = *execprog
		} else {
			// Execute specified file.
			filename = flag.Arg(0)
		}
		thread.Name = "exec " + filename
		globals, err = starlark.ExecFile(thread, filename, src, predecls)
		if err != nil {
			repl.PrintError(err)
			return 1
		}
	case flag.NArg() == 0:
		thread.Name = "REPL"
		repl.REPL(thread, globals)
	default:
		log.Print("want at most one file name")
		return 1
	}

	if *showenv {
		for _, name := range globals.Keys() {
			if !strings.HasPrefix(name, "_") {
				fmt.Fprintf(os.Stderr, "%s = %s\n", name, globals[name])
			}
		}
	}

	if reporter.count > 0 {
		return 10
	}

	return 0
}
