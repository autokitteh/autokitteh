package systest

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	jd "github.com/josephburnett/jd/lib"

	"go.autokitteh.dev/autokitteh/tests"
)

func runCheck(t *testing.T, step string, ak *tests.AKResult, resp *httpResponse) error {
	match := steps.FindStringSubmatch(step)
	switch match[1] {
	case "output":
		return checkAKOutput(t, step, ak)
	case "return":
		return checkAKReturnCode(step, ak)
	case "resp":
		return checkHTTPResponse(step, resp)
	case "capture_jq":
		return captureJQ(t, step, ak, resp)
	case "capture_re":
		return captureRE(t, step, ak, resp)
	default:
		return errors.New("unhandled check")
	}
}

// 'meow world' and something -> "meow world", " and something"
// "meow world" and something -> "meow world", " and something"
// meow world and something -> "meow", " world and something"
func nextField(text string) (string, string) {
	if len(text) == 0 {
		return "", ""
	}

	if text[0] == '\'' || text[0] == '"' {
		i := strings.IndexFunc(text[1:], func(r rune) bool { return r == rune(text[0]) })
		return text[1 : i+1], text[i+2:]
	}

	a, b, _ := strings.Cut(text, " ")
	return a, b
}

func checkAKOutput(t *testing.T, step string, ak *tests.AKResult) error {
	match := akCheckOutput.FindStringSubmatch(step)
	want := strings.TrimSpace(match[3])
	want = strings.TrimPrefix(want, "'")
	want = strings.TrimSuffix(want, "'")
	got := ak.Output

	if strings.HasPrefix(match[2], "file") {
		b, err := os.ReadFile(want)
		if err != nil {
			return fmt.Errorf("failed to read embedded file: %w", err)
		}
		want = strings.TrimSpace(string(b))
	}

	t.Logf("step: %q\nwant: %q\ngot: %q", step, want, got)

	switch match[1] {
	case "equals_json":
		wantJSON, err := jd.ReadJsonString(want)
		if err != nil {
			return fmt.Errorf("failed to parse expected JSON: %w\n%s", err, want)
		}

		gotJSON, err := jd.ReadJsonString(got)
		if err != nil {
			return fmt.Errorf("failed to parse actual JSON: %w\n%s", err, got)
		}

		diff := wantJSON.Diff(gotJSON)
		if len(diff) > 0 {
			return jsonCheckFailed(diff)
		}
	case "equals":
		if want != got {
			return stringCheckFailed(want, got)
		}
	case "contains_json":
		// Using a library like go-cmp
		opts := cmpopts.IgnoreFields(struct {
			Build struct{ Info struct{ Memo string } }
		}{}, "Build.Info.Memo")
		if diff := cmp.Diff(want, got, opts); diff != "" {
			return fmt.Errorf("expected JSON not found in actual output:\n%s", diff)
		}
		return nil
	case "contains":
		if len(want) == 0 && got != want { // Empty string is always contained.
			return stringCheckFailed(want, got)
		}
		if !strings.Contains(got, want) {
			return stringCheckFailed(want, got)
		}
	case "regex":
		matched, err := regexp.MatchString(want, got)
		if err != nil {
			return fmt.Errorf("failed to match regex: %w", err)
		}
		if !matched {
			return stringCheckFailed(want, got)
		}
	case "equals_jq":
		q, expected := nextField(want)

		got, err := jq(ak.Output, q)
		if err != nil {
			return fmt.Errorf("failed to run jq: %w", err)
		}

		if expected != got {
			return stringCheckFailed(want, got)
		}

	default:
		return errors.New("unhandled AK check type")
	}
	return nil
}

func checkAKReturnCode(step string, ak *tests.AKResult) error {
	match := akCheckReturn.FindStringSubmatch(step)
	expected, err := strconv.Atoi(match[1])
	if err != nil {
		return fmt.Errorf("failed to parse expected return code: %w", err)
	}
	if expected != ak.ReturnCode {
		msg := fmt.Sprintf("got return code %d, want %d", ak.ReturnCode, expected)
		// Append the AK output for context, if there is any.
		if ak.Output != "" {
			msg += "\n" + ak.Output
		}
		return errors.New(msg)
	}
	return nil
}

func checkHTTPResponse(step string, resp *httpResponse) error {
	match := httpChecks.FindStringSubmatch(step)
	switch match[1] {
	case "body", "redirect":
		return checkHTTPResponseBody(step, resp)
	case "code":
		return checkHTTPStatusCode(step, resp)
	default:
		return errors.New("unhandled HTTP check type")
	}
}

func checkHTTPResponseBody(step string, resp *httpResponse) error {
	return errors.New("not implemented yet")
}

func checkHTTPStatusCode(step string, resp *httpResponse) error {
	match := httpCheckStatus.FindStringSubmatch(step)
	expected, err := strconv.Atoi(match[1])
	if err != nil {
		return fmt.Errorf("failed to parse expected return code: %w", err)
	}
	if expected != resp.resp.StatusCode {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("got return code %d, want %d", resp.resp.StatusCode, expected))
		// Append the response for context, if there is any.
		for k, v := range resp.resp.Header {
			sb.WriteString(fmt.Sprintf("\n%s: %s", k, strings.Join(v, ", ")))
		}
		if resp.body != "" {
			sb.WriteString("\n" + resp.body)
		}
		return errors.New(sb.String())
	}
	return nil
}

func stringCheckFailed(want, got string) error {
	edits := myers.ComputeEdits(span.URIFromPath("want"), want+"\n", got+"\n")
	return errors.New(fmt.Sprint("\n", gotextdiff.ToUnified("want", "got", want+"\n", edits)))
}

func jsonCheckFailed(diff jd.Diff) error {
	return errors.New(diff.Render())
}
