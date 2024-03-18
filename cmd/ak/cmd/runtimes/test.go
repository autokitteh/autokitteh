package runtimes

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/spf13/cobra"
	"golang.org/x/tools/txtar"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkbuildfile"
)

var testCmd = common.StandardCommand(&cobra.Command{
	Use:     "test <txtar-path> [--timeout <t>] [--quiet]",
	Short:   "Test a program",
	Aliases: []string{"t"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		f, err := os.OpenFile(args[0], os.O_RDONLY, 0)
		if err != nil {
			return fmt.Errorf("open: %w", err)
		}

		defer f.Close()

		var b *sdkbuildfile.BuildFile

		var expectedPrints string

		bs, err := io.ReadAll(f)
		if err != nil {
			return err
		}

		a := txtar.Parse(bs)
		if len(a.Files) == 0 {
			return fmt.Errorf("empty txtar archive")
		}

		expectedPrints = strings.TrimRight(string(a.Comment), "\n")

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

		vs, prints, err := run(cmd.Context(), b, path)
		if err != nil {
			return err
		}

		if emitValues {
			common.RenderKV("values", vs)
		}

		actual := strings.Join(prints, "\n")
		if actual != expectedPrints {
			edits := myers.ComputeEdits(span.URIFromPath("want"), expectedPrints, actual)
			return errors.New(fmt.Sprint(gotextdiff.ToUnified("want", "got", expectedPrints, edits)))
		}

		return nil
	},
})

func init() {
	testCmd.Flags().DurationVarP(&tmo, "timeout", "t", 0, "timeout")
	testCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "do not print anything but errors")
}
