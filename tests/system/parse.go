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
	steps = regexp.MustCompile(`^(?:(ak|http|output|req|resp|return|wait|setenv|capture_jq|capture_re|user)\s)|(exit)$`)

	// ak *
	// http <get|post> *
	actions = regexp.MustCompile(`^(ak|http\s+(get|post)|wait|setenv|user)\s+(.+)`)
	// wait <duration> for session <session ID>
	waitAction = regexp.MustCompile(`^wait\s+(.+)\s+(for|unless)\s+session\s+(.+)`)

	// capture_jq <name> <jq expression>
	jqCheck = regexp.MustCompile(`^capture_jq\s+(\w+)\s+(.+)`)

	// capture_re <name> <regexp>
	reCheck = regexp.MustCompile(`^capture_re\s+(\w+)\s+(.+)`)

	// output <equals|equals_json|contains|regex> [file] *
	akCheckOutput = regexp.MustCompile(`^output\s+(equals|equals_json|contains|regex|equals_jq)\s+(file\s+)?(.+|'.*')`)
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
	// Extra up configuration options.
	Server map[string]any `json:"server" yaml:"server"`

	// If set, only this test will run.
	Exclusive bool `json:"exclusive" yaml:"exclusive"`

	// General config for the test itself.
	AK akTestConfig `json:"ak" yaml:"ak"`
}

func useTempDir(t *testing.T) {
	td := t.TempDir()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	})

	if err := os.Chdir(td); err != nil {
		t.Fatalf("failed to switch to temporary directory: %v", err)
	}

	// Don't use the user's "config.yaml" file, it may violate isolation
	// by forcing tests to use shared and/or persistent resources.
	t.Setenv("XDG_CONFIG_HOME", td)
	t.Setenv("XDG_DATA_HOME", td)
}

func writeEmbeddedFiles(t *testing.T, fs []txtar.File) {
	for _, f := range fs {
		if err := os.MkdirAll(filepath.Dir(f.Name), tempDirPerm); err != nil {
			t.Fatalf("failed to create directory for embedded file %q: %v", f.Name, err)
		}

		if err := os.WriteFile(f.Name, f.Data, tempFilePerm); err != nil {
			t.Fatalf("failed to write embedded file %q: %v", f.Name, err)
		}
	}
}
