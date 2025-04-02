package sessions

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"regexp"
	"strings"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/spf13/cobra"
	"golang.org/x/tools/txtar"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	// /tmp/ak-user-2767870919/main.py:6.1,main
	// can't load main from /tmp/ak-user-1368663929 - error: No module named 'garfield_loves_pizza'
	userRe = regexp.MustCompile(`/.*/ak-user-.*?[/ ]`)
	// runner/main.py:6.1,main, in _call
	runnerRe = regexp.MustCompile(`.*runner.*/.*\.py.*`)

	// File "/opt/hostedtoolcache/Python/3.12.8/x64/lib/python3.12/concurrent/futures/_base.py", line 401, in __get_result`
	pyLibRe = regexp.MustCompile(`File ".*/lib/python3\.\d+/(.*\.py)", line (\d+), in (.*)`)

	// Some python version like to put an annoying ^^^^ marker to show where the error is.
	pyAnnoyingErrorLocationMarkerRe = regexp.MustCompile(`^\s*~*\^+$`)
)

func normalizePath(p string) string {
	// Remove location specific prefix of Python standard library.
	line := pyLibRe.ReplaceAllString(p, `py-lib/$1, line XXX, in $3`)
	if line != p {
		return line
	}

	// Too many changes in runner, just show runner
	if runnerRe.MatchString(p) {
		return "   ak-runner"
	}

	// Remove /tmp/ak-userXXX prefix.
	return userRe.ReplaceAllString(p, "")
}

func normalizeOutput(ps string) (string, error) {
	var buf strings.Builder

	s := bufio.NewScanner(strings.NewReader(ps))
	for s.Scan() {
		line := normalizePath(s.Text())

		if pyAnnoyingErrorLocationMarkerRe.MatchString(line) {
			continue
		}

		buf.WriteString(line)
		buf.WriteRune('\n')
	}

	if err := s.Err(); err != nil {
		return "", fmt.Errorf("normalize: %w", err)
	}

	return strings.TrimSpace(buf.String()), nil
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

		txtarFS, err := kittehs.TxtarToFS(a)
		if err != nil {
			return fmt.Errorf("internal error: %w", err)
		}

		if !bid.IsValid() {
			b, err := common.Build(common.Client().Runtimes(), txtarFS, nil, nil)
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

		expectedCallsTxt, err := fs.ReadFile(txtarFS, "calls.txt")
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("open calls.txt: %w", err)
		}

		expectedErrTxt, err := fs.ReadFile(txtarFS, "error.txt")
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("open error.txt: %w", err)
		}

		s := sdktypes.NewSession(bid, ep, nil, nil).WithDeploymentID(did).WithProjectID(pid)

		ctx, cancel := common.LimitedContext()
		defer cancel()

		sid, err := sessions().Start(ctx, s)
		if err != nil {
			return fmt.Errorf("start session: %w", err)
		}
		pageSize = 10
		if _, err := sessionWatch(sid, sdktypes.SessionStateTypeUnspecified, ""); err != nil {
			return err
		}

		rs, err := sessions().GetPrints(ctx, sid, sdktypes.PaginationRequest{
			Ascending: true,
		})
		if err != nil {
			return err
		}

		prints := make([]string, 0, len(rs.Prints))
		for _, r := range rs.Prints {
			ps, err := r.Value.ToString()
			if err != nil {
				ps = ""
			}

			ps, err = normalizeOutput(ps)
			if err != nil {
				return err
			}
			prints = append(prints, ps)
		}

		expected := strings.TrimSpace(string(a.Comment))
		actual := strings.TrimSpace(strings.Join(prints, "\n"))

		var errs []error

		if actual != expected {
			edits := myers.ComputeEdits(span.URIFromPath("want"), expected, actual)
			errs = append(errs, errors.New(fmt.Sprint("output:\n", gotextdiff.ToUnified("want", "got", expected, edits))))
		}

		if expectedErrTxt != nil {
			results, err := sessions().GetLog(ctx, sdkservices.SessionLogRecordsFilter{
				SessionID: sid,
				Types:     sdktypes.StateSessionLogRecordType,
				PaginationRequest: sdktypes.PaginationRequest{
					PageSize: 1,
				},
			})
			if err != nil {
				return fmt.Errorf("get calls: %w", err)
			}

			var actual string

			if len(results.Records) != 0 {
				errState := results.Records[0].GetState().GetError()
				if errState.IsValid() {
					actual = strings.TrimSpace(errState.GetProgramError().Value().GetString().Value())
				}
			}

			expectedErrTxt, err := normalizeOutput(string(expectedErrTxt))
			if err != nil {
				return err
			}
			actual, err = normalizeOutput(actual)
			if err != nil {
				return err
			}
			if err := testDiff("error", expectedErrTxt, actual); err != nil {
				errs = append(errs, err)
			}
		}

		if expectedCallsTxt != nil {
			results, err := sessions().GetLog(ctx, sdkservices.SessionLogRecordsFilter{
				SessionID: sid,
				Types:     sdktypes.CallSpecSessionLogRecordType,
				PaginationRequest: sdktypes.PaginationRequest{
					Ascending: true,
				},
			})
			if err != nil {
				return fmt.Errorf("get calls: %w", err)
			}

			var callsTxt strings.Builder
			for _, r := range results.Records {
				f, _, _ := r.GetCallSpec().Data()
				fmt.Fprintf(&callsTxt, "%s\n", f.GetFunction().Name())
			}

			if err := testDiff("calls", kittehs.StringWithoutComments(string(expectedCallsTxt)), callsTxt.String()); err != nil {
				errs = append(errs, err)
			}
		}

		return errors.Join(errs...)
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

func testDiff(what, expected, actual string) error {
	expected = strings.TrimSpace(expected)
	actual = strings.TrimSpace(actual)

	if expected == actual {
		return nil
	}

	edits := myers.ComputeEdits(span.URIFromPath("want"), expected, actual)
	diff := gotextdiff.ToUnified("want", "got", expected, edits)
	return fmt.Errorf("%s:\n%s", what, diff)
}
