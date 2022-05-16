package cmdupdate

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"

	T "github.com/autokitteh/autokitteh/cmd/ak/clitools"
	P "github.com/autokitteh/autokitteh/cmd/ak/cmdproject/projecttools"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiprogram"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiproject"
)

var (
	opts struct {
		name, path string
		memo       cli.StringSlice
		disable    bool
	}

	Cmd = cli.Command{
		Name:    "update",
		Aliases: []string{"u"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "name",
				Aliases:     []string{"n"},
				Destination: &opts.name,
			},
			&cli.StringFlag{
				Name:        "path",
				Aliases:     []string{"p"},
				Destination: &opts.path,
			},
			&cli.StringSliceFlag{
				Name:        "memo",
				Aliases:     []string{"m"},
				Destination: &opts.memo,
			},
			&cli.BoolFlag{
				Name:        "disable",
				Aliases:     []string{"d"},
				Destination: &opts.disable,
			},
		},
		Action: func(c *cli.Context) error {
			args := c.Args().Slice()

			if len(args) != 1 {
				return fmt.Errorf("single project id expected")
			}

			memo := make(map[string]string, len(opts.memo.Value()))
			for _, kv := range opts.memo.Value() {
				parts := strings.SplitN(kv, "=", 2)
				var v string
				if len(parts) > 1 {
					v = parts[1]
				}
				memo[parts[0]] = v
			}

			var path *apiprogram.Path
			if opts.path != "" {
				var err error
				if path, err = apiprogram.ParsePathString(opts.path); err != nil {
					return err
				}
			}

			data := (&apiproject.ProjectData{}).SetName(opts.name).SetMemo(memo).SetMainPath(path).SetEnabled(!opts.disable)

			if err := P.Projects().Update(
				T.Context,
				apiproject.ProjectID(args[0]),
				data,
			); err != nil {
				return err
			}

			return nil
		},
	}
)
