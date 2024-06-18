package sessions

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/spf13/cobra"
	"golang.org/x/tools/txtar"
	"gopkg.in/yaml.v3"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var testCmd = common.StandardCommand(&cobra.Command{
	Use:   "test <txtar-files> [--build-id=...] [--env=...] [--deployment-id=...] [--entrypoint=...] [--quiet] [--timeout DURATION] [--poll-interval DURATION] [--no-timestamps]",
	Short: "Test a session run",
	Args:  cobra.MinimumNArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		did, eid, bid, ep, inputs, err := sessionArgs()
		if err != nil {
			return err
		}

		var a *txtar.Archive

		for _, path := range args {
			f, err := os.OpenFile(path, os.O_RDONLY, 0)
			if err != nil {
				return fmt.Errorf("open: %w", err)
			}

			defer f.Close()

			bs, err := io.ReadAll(f)
			if err != nil {
				return fmt.Errorf("read: %w", err)
			}

			a1 := txtar.Parse(bs)
			if a == nil {
				// comment comes only from the first file.
				a = &txtar.Archive{Comment: a1.Comment, Files: a1.Files}
			} else {
				a.Files = append(a.Files, a1.Files...)
			}
		}

		if a == nil || len(a.Files) == 0 {
			return fmt.Errorf("empty txtar archive")
		}

		// inputs from file overwrite inputs from command line.
		for i, f := range a.Files {
			if f.Name != "inputs.yaml" {
				continue
			}

			var data map[string]any

			if err := yaml.Unmarshal(f.Data, &data); err != nil {
				return fmt.Errorf("unmarshal inputs: %w", err)
			}

			for k, v := range data {
				if inputs[k], err = sdktypes.WrapValue(v); err != nil {
					return fmt.Errorf("wrap input value: %w", err)
				}
			}

			a.Files = append(a.Files[:i], a.Files[i+1:]...)
			break
		}

		if !ep.IsValid() {
			if len(a.Files) == 0 {
				return fmt.Errorf("no entrypoint specified and no files found in txtar archive")
			}

			// txtar coloring in vscode doesn't like ':', so replace it with a space.
			a.Files[0].Name = strings.ReplaceAll(a.Files[0].Name, " ", ":")

			if ep, err = sdktypes.StrictParseCodeLocation(a.Files[0].Name); err != nil {
				return fmt.Errorf("invalid entrypoint: %w", err)
			}

			// Filter out the non-path from the path.
			a.Files[0].Name = ep.Path()
		}

		fs, err := kittehs.TxtarToFS(a)
		if err != nil {
			return fmt.Errorf("internal error: %w", err)
		}

		if !bid.IsValid() {
			b, err := common.Build(common.Client().Runtimes(), fs, nil, nil)
			if err != nil {
				return err
			}

			var buf bytes.Buffer
			if err := b.Write(&buf); err != nil {
				return fmt.Errorf("write build: %w", err)
			}

			ctx, cancel := common.LimitedContext()
			defer cancel()

			if bid, err = common.Client().Builds().Save(ctx, sdktypes.NewBuild(), buf.Bytes()); err != nil {
				return fmt.Errorf("save build: %w", err)
			}
		}

		s := sdktypes.NewSession(bid, ep, inputs, nil).WithDeploymentID(did).WithEnvID(eid)

		ctx, cancel := common.LimitedContext()
		defer cancel()

		sid, err := sessions().Start(ctx, s)
		if err != nil {
			return fmt.Errorf("start session: %w", err)
		}

		rs, err := sessionWatch(sid, sdktypes.SessionStateTypeUnspecified)
		if err != nil {
			return err
		}

		var prints strings.Builder
		for _, r := range rs {
			if p, ok := r.GetPrint(); ok {
				prints.WriteString(p)
				prints.WriteRune('\n')
			}
		}

		expected := strings.TrimSpace(string(a.Comment))
		actual := strings.TrimSpace(prints.String())

		if actual != expected {
			edits := myers.ComputeEdits(span.URIFromPath("want"), expected, actual)
			return errors.New(fmt.Sprint(gotextdiff.ToUnified("want", "got", expected, edits)))
		}

		return nil
	},
})

func init() {
	testCmd.Flags().DurationVarP(&watchTimeout, "timeout", "t", 0, "watch timeout duration")
	testCmd.Flags().BoolVar(&noTimestamps, "no-timestamps", false, "omit timestamps from watch output")
	testCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "don't print anything, just wait to finish")
	testCmd.Flags().StringArrayVarP(&inputs, "input", "I", nil, `zero or more "key=value" pairs, where value is a JSON value`)
}
