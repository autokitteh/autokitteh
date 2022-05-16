package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/cosiner/argv"
	"github.com/urfave/cli/v2"
)

var cmdSh = cli.Command{
	Name:    "shell",
	Aliases: []string{"sh"},
	Action: func(c *cli.Context) error {
		args := c.Args().Slice()

		if len(args) > 1 {
			return fmt.Errorf("expecting no arguments or a file name")
		}

		var read func() (string, error)

		if len(args) == 0 {
			l, err := readline.NewEx(&readline.Config{
				Prompt:          "\033[31m»\033[0m ",
				InterruptPrompt: "^C",
				EOFPrompt:       "exit",

				HistoryFile:       "/tmp/.aksh.history",
				HistorySearchFold: true,
			})
			if err != nil {
				return fmt.Errorf("readline: %w", err)
			}

			defer l.Close()

			read = l.Readline
		} else {
			f, err := os.Open(args[0])
			if err != nil {
				return fmt.Errorf("open %q: %w", args[0], err)
			}

			defer f.Close()
			s := bufio.NewScanner(f)

			read = func() (string, error) {
				if !s.Scan() {
					return "", io.EOF
				}

				return s.Text(), s.Err()
			}
		}

		cli.OsExiter = func(int) {}

		for {
			text, err := read()
			if err == readline.ErrInterrupt {
				if len(text) == 0 {
					break
				} else {
					continue
				}
			} else if err == io.EOF {
				break
			}

			text = strings.TrimSpace(text)

			if text == "" {
				continue
			}

			args, err := argv.Argv(os.Args[0]+" "+text, nil, nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v", err.Error())
				continue
			}

			if len(args) > 1 {
				fmt.Fprintf(os.Stderr, "no pipes allowed")
				continue
			}

			fmt.Fprintf(os.Stderr, "exec: %v\n", args[0])

			rc := run(newApp(), args[0])

			fmt.Fprintf(os.Stderr, "rc: %d\n", rc)
		}

		return nil
	},
}
