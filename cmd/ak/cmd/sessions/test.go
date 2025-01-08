package sessions

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/spf13/cobra"
	"golang.org/x/tools/txtar"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	// /tmp/ak-user-2767870919/main.py:6.1,main
	trimRe = regexp.MustCompile(`\/.*\/ak-(user|runner)-.*?\/`)
	//  File "runner/main.py", line 323, in _call
	runnerRe = regexp.MustCompile(`(ak-)?runner(-\d+)?/.*py`)
)

func normalizePath(p string) string {

	// Remove location specific prefix of Python standard library.
	const pyLibPrefix = "/lib/python"
	i := strings.Index(p, pyLibPrefix)
	if i != -1 {
		p = p[i+len(pyLibPrefix):]

		j := strings.Index(p, "/")
		if j > 0 {
			p = "py/" + p[j+1:]
		}

		return p
	}

	// Remove ak-runner and ak-user.
	return trimRe.ReplaceAllString(p, "")
}

var testCmd = common.StandardCommand(&cobra.Command{
	Use:   "test <txtar-file> [--build-id=...] [--project project] [--deployment-id=...] [--entrypoint=...] [--quiet] [--timeout DURATION] [--poll-interval DURATION] [--no-timestamps]",
	Short: "Test a session run",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		did, pid, bid, ep, err := sessionArgs()
		if err != nil {
			return err
		}

		f, err := os.OpenFile(args[0], os.O_RDONLY, 0)
		if err != nil {
			return fmt.Errorf("open: %w", err)
		}

		defer f.Close()

		bs, err := io.ReadAll(f)
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}

		a := txtar.Parse(bs)
		if len(a.Files) == 0 {
			return errors.New("empty txtar archive")
		}

		if !ep.IsValid() {
			if len(a.Files) == 0 {
				return errors.New("no entrypoint specified and no files found in txtar archive")
			}

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

			if bid, err = common.Client().Builds().Save(ctx, sdktypes.NewBuild().WithProjectID(pid), buf.Bytes()); err != nil {
				return fmt.Errorf("save build: %w", err)
			}
		}

		s := sdktypes.NewSession(bid, ep, nil, nil).WithDeploymentID(did).WithProjectID(pid)

		ctx, cancel := common.LimitedContext()
		defer cancel()

		sid, err := sessions().Start(ctx, s)
		if err != nil {
			return fmt.Errorf("start session: %w", err)
		}
		pageSize = 10
		rs, err := sessionWatch(sid, sdktypes.SessionStateTypeUnspecified, "")
		if err != nil {
			return err
		}

		slices.SortFunc(rs, func(a, b sdktypes.SessionLogRecord) int {
			return a.Timestamp().Compare(b.Timestamp())
		})

		var prints strings.Builder

		for _, r := range rs {
			if p, ok := r.GetPrint(); ok {
				p = normalizePath(p)

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
	testCmd.Flags().StringVarP(&project, "project", "p", "", "project name or ID")
	testCmd.Flags().StringVarP(&deploymentID, "deployment-id", "d", "", "deployment ID")
	testCmd.Flags().StringVarP(&buildID, "build-id", "b", "", "build ID, mutually exclusive with --deployment-id")
}
