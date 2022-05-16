package cmdcreate

import (
	"strings"

	"github.com/urfave/cli/v2"

	T "github.com/autokitteh/autokitteh/cmd/ak/clitools"
	P "github.com/autokitteh/autokitteh/cmd/ak/cmdproject/projecttools"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiaccount"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiprogram"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiproject"
	"github.com/autokitteh/autokitteh/internal/pkg/projectsstore"
)

var (
	opts struct {
		name, aid, path string
		memo            cli.StringSlice
	}

	Cmd = cli.Command{
		Name:    "create",
		Aliases: []string{"c"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "account-id",
				Aliases:     []string{"a", "aid"},
				Destination: &opts.aid,
			},
			&cli.StringFlag{
				Name:        "name",
				Aliases:     []string{"n"},
				Destination: &opts.name,
			},
			&cli.StringSliceFlag{
				Name:        "memo",
				Aliases:     []string{"m"},
				Destination: &opts.memo,
			},
			&cli.StringFlag{
				Name:        "path",
				Aliases:     []string{"p"},
				Destination: &opts.path,
			},
		},
		Action: func(c *cli.Context) error {
			memo := make(map[string]string, len(opts.memo.Value()))
			for _, kv := range opts.memo.Value() {
				parts := strings.SplitN(kv, "=", 2)
				var v string
				if len(parts) > 1 {
					v = parts[1]
				}
				memo[parts[0]] = v
			}

			path, err := apiprogram.ParsePathString(opts.path)
			if err != nil {
				return err
			}

			data := (&apiproject.ProjectData{}).SetMemo(memo).SetName(opts.name).SetMainPath(path)

			id, err := P.Projects().Create(
				T.Context,
				apiaccount.AccountID(opts.aid),
				projectsstore.AutoProjectID,
				data,
			)

			if err != nil {
				return err
			}

			T.Show(
				map[string]string{
					"id": id.String(),
				},
			)

			return nil
		},
	}
)
