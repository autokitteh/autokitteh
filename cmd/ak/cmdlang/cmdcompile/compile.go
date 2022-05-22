package cmdcompile

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/urfave/cli/v2"
	"google.golang.org/protobuf/proto"

	T "github.com/autokitteh/autokitteh/cmd/ak/clitools"
	L "github.com/autokitteh/autokitteh/cmd/ak/cmdlang/langtools"
	"github.com/autokitteh/autokitteh/sdk/api/apiprogram"
	"github.com/autokitteh/autokitteh/internal/pkg/lang/langtools"
)

// TODO: prevent different files might end up in the same output file (x.starlark.ak, x.starlark -> x.akm). Maybe allow multiple modules in a single file?
// TODO: more configurable.

var (
	flags struct {
		predecls cli.StringSlice
	}

	Cmd = cli.Command{
		Name:    "compile",
		Aliases: []string{"c"},
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:        "predecl",
				Aliases:     []string{"p"},
				Destination: &flags.predecls,
			},
		},
		Action: func(c *cli.Context) error {
			for _, arg := range c.Args().Slice() {
				T.Infof("compiling %s", arg)
				if err := compileFile(arg, flags.predecls.Value()); err != nil {
					return fmt.Errorf("%s: %w", arg, err)
				}
			}

			return nil
		},
	}
)

func compileFile(path string, predecls []string) error {
	ctx := T.Context

	src, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}

	lpath, err := apiprogram.NewPath("", path, "")
	if err != nil {
		return err
	}

	mod, ext, err := langtools.CompileModule(ctx, L.Catalog(), predecls, lpath, src)
	if err != nil {
		return err
	}

	var dst string

	if ext == "" {
		dst = path + ".akm"
	} else {
		dst = strings.TrimSuffix(path, ext) + "akm"
	}

	compiled, err := proto.Marshal(mod.PB())
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if err := ioutil.WriteFile(dst, compiled, 0644); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}
