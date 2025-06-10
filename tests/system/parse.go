package systest

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"golang.org/x/tools/txtar"
)

const (
	tempDirPerm  = 0o755 // drwxr-xr-x
	tempFilePerm = 0o644 // -rw-r--r--
)

var (
	steps = regexp.MustCompile(`^(?:(ak|http|output|req|resp|return|wait|setenv|capture_jq|capture_re|user|file|exec)\s)|(exit)$`)

	// ak *
	// http <get|post> *
	actions = regexp.MustCompile(`^(ak|http\s+(get|post)|wait|setenv|user|exec)\s+(.+)`)
	// wait <duration> for session <session ID>
	waitAction = regexp.MustCompile(`^wait\s+(.+)\s+(for|unless)\s+session\s+(.+)`)

	// capture_jq <name> <jq expression>
	jqCheck = regexp.MustCompile(`^capture_jq\s+(\w+)\s+(.+)`)

	// capture_re <name> <regexp>
	reCheck = regexp.MustCompile(`^capture_re\s+(\w+)\s+(.+)`)

	// output <equals|equals_json|contains|regex> [file] *
	akCheckOutput = regexp.MustCompile(`^output\s+(equals|equals_json|contains|regex|equals_jq)\s+(file\s+)?(.+|'.*')`)

	// file <filename> contains <text>
	fileChecks = regexp.MustCompile(`^file\s+(.+)\s+contains\s+(.+)`)

	// return code == <int>
	akCheckReturn = regexp.MustCompile(`^return\s+code\s*==\s*(\d+)$`)

	// req header <name> = <value>
	httpCustomHeader = regexp.MustCompile(`^req\s+header\s+([\w-]+)\s*=\s*(.+)`)
	// req body [file] *
	httpCustomBody = regexp.MustCompile(`^req\s+body\s+(file\s+)?(.+)`)

	httpChecks = regexp.MustCompile(`^resp\s+(body|redirect|code|header)`)
	// resp <body|redirect> <equals|contains|regex> [file] *
	httpCheckOutput = regexp.MustCompile(`^resp\s+(body|redirect)\s+(equals|contains|regex)\s+(file\s+)?(.+|'.*')`)
	// resp code == <int>
	httpCheckStatus = regexp.MustCompile(`^resp\s+code\s*==\s*(\d+)$`)
	// resp header <name> == <value>
	httpCheckHeader = regexp.MustCompile(`^resp\s+header\s+([\w-]+)\s*==\s*(.+)`)
)

type akTestConfig struct {
	// Extra arguments to pass to every AK command.
	ExtraArgs []string `json:"extra_args" yaml:"extra_args"`
}

type testConfig struct {
	// Extra "ak up" configuration options.
	Server map[string]any `json:"server" yaml:"server"`

	// If set, only this test will run.
	// Same as the "-run" flag in "go test", but easier to use.
	Exclusive bool `json:"exclusive" yaml:"exclusive"`

	// General config for the test itself.
	AK akTestConfig `json:"ak" yaml:"ak"`
}

func writeEmbeddedFiles(t *testing.T, fs []txtar.File) {
	if err := os.Mkdir("archive", tempDirPerm); err != nil {
		t.Fatal("failed to create directory 'archive':", err)
	}

	if err := os.Chdir("archive"); err != nil {
		t.Fatal("failed to change working directory to 'archive':", err)
	}

	for _, f := range fs {
		if err := os.MkdirAll(filepath.Dir(f.Name), tempDirPerm); err != nil {
			t.Fatalf("failed to create directory for embedded file %q: %v", f.Name, err)
		}

		if err := os.WriteFile(f.Name, f.Data, tempFilePerm); err != nil {
			t.Fatalf("failed to write embedded file %q: %v", f.Name, err)
		}
	}
}
