package runtimes

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/tools/txtar"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkbuildfile"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	path       string
	tmo        time.Duration
	txtarFile  bool
	emitValues bool
	quiet      bool
)

var runCmd = common.StandardCommand(&cobra.Command{
	Use:     "run <build file|program file> [--txtar] [--path path] [-timeout t] [--values] [--test] [--quiet]",
	Short:   `Run a program`,
	Aliases: []string{"r"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		f, err := os.OpenFile(args[0], os.O_RDONLY, 0)
		if err != nil {
			return fmt.Errorf("open: %w", err)
		}

		defer f.Close()

		var b *sdkbuildfile.BuildFile

		if txtarFile {
			bs, err := io.ReadAll(f)
			if err != nil {
				return fmt.Errorf("read: %w", err)
			}

			a := txtar.Parse(bs)
			if len(a.Files) == 0 {
				return fmt.Errorf("empty txtar archive")
			}

			fs, err := kittehs.TxtarToFS(a)
			if err != nil {
				return fmt.Errorf("internal error: %w", err)
			}

			if b, err = common.Build(runtimes(), fs, nil, nil); err != nil {
				return err
			}

			if path == "" {
				path = filepath.Clean(a.Files[0].Name)
			}
		} else if isBuild, err := isBuildFile(f); err != nil {
			return err
		} else if isBuild {
			var err error
			if b, err = sdkbuildfile.Read(f); err != nil {
				return fmt.Errorf("read build file: %w", err)
			}
		} else {
			// single program file
			if path != "" {
				return errors.New("cannot specify path with single program file")
			}

			path = filepath.Clean(args[0])

			var buildPaths []string

			if !local {
				// if builds is not local, we don't want to upload the entrie
				// local directory. This, however, prevents single files from being
				// run remotely if the load other files.
				buildPaths = []string{path}
			}

			var err error
			if b, err = common.Build(runtimes(), os.DirFS("."), buildPaths, nil); err != nil {
				return err
			}
		}

		vs, _, err := run(cmd.Context(), b, path)
		if err != nil {
			return err
		}

		if emitValues {
			common.RenderKV("values", vs)
		}

		return nil
	},
})

func init() {
	runCmd.Flags().StringVarP(&path, "path", "p", "", "entrypoint path")
	runCmd.Flags().DurationVarP(&tmo, "timeout", "t", 0, "timeout")
	runCmd.Flags().BoolVar(&txtarFile, "txtar", false, "input file is a txtar archive containing program")
	runCmd.Flags().BoolVarP(&emitValues, "values", "v", false, "emit result values")
	runCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "do not print anything but errors")
}

func run(ctx context.Context, b *sdkbuildfile.BuildFile, path string) (map[string]sdktypes.Value, []string, error) {
	var prints []string

	cbs := &sdkservices.RunCallbacks{
		Print: func(_ context.Context, _ sdktypes.RunID, msg string) {
			if !quiet {
				fmt.Println(msg)
			}
			prints = append(prints, msg)
		},
	}

	if tmo > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, tmo)
		defer cancel()
	}

	run, err := runtimes().Run(ctx, sdktypes.NewRunID(), sdktypes.InvalidSessionID, path, b, nil, cbs)
	if err != nil {
		return nil, nil, fmt.Errorf("run build: %w", err)
	}

	run.Close()

	return run.Values(), prints, nil
}

func isBuildFile(f *os.File) (bool, error) {
	if !sdkbuildfile.IsBuildFile(f) {
		return false, nil
	}

	if _, err := f.Seek(0, 0); err != nil {
		return false, fmt.Errorf("seek: %w", err)
	}

	return true, nil
}
