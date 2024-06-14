package systest

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"golang.org/x/tools/txtar"
)

const (
	tempDirPerm  = 0o755 // drwxr-xr-x
	tempFilePerm = 0o644 // -rw-r--r--
)

var (
	steps = regexp.MustCompile(`^(ak|http|output|req|resp|return|wait|setenv)\s`)

	// ak *
	// http <get|post> *
	actions = regexp.MustCompile(`^(ak|http\s+(get|post)|wait|setenv)\s+(.+)`)
	// wait <duration> for session <session ID>
	waitAction = regexp.MustCompile(`^wait\s+(.+)\s+(for|unless)\s+session\s+(.+)`)

	// output <equals|equals_json|equals_contains|regex> [file] *
	akCheckOutput = regexp.MustCompile(`^output\s+(equals|equals_json|contains|regex)\s+(file\s+)?(.+|'.*')`)
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

func readTestFile(t *testing.T, path string) []string {
	a, err := txtar.ParseFile(path)
	if err != nil {
		t.Fatalf("failed to load file: %v", err)
	}

	if len(a.Comment) == 0 {
		t.Fatalf("nothing to do in %q: txtar comment section is empty", path)
	}

	for i, f := range a.Files {
		// Support embedded txtars.
		if filepath.Ext(f.Name) == ".txtar" {
			a.Files[i].Data = []byte(strings.ReplaceAll(string(f.Data), "~~", "--"))
		}
	}

	useTempDir(t)
	writeEmbeddedFiles(t, a.Files)
	return parseTestFile(t, a)
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

func parseTestFile(t *testing.T, a *txtar.Archive) []string {
	lines := strings.Split(string(a.Comment), "\n")
	errors := 0
	for i, line := range lines {
		// Trim redundant whitespaces and single-line comments
		// (but don't discard empty lines, to preserve line numbers).
		lines[i] = strings.TrimSpace(line)
		lines[i] = regexp.MustCompile(`^\s*#.*`).ReplaceAllString(line, "")
		line = lines[i]

		if line == "" {
			continue
		}

		// (Eventually) fail on unrecognized or invalid steps.
		match := steps.FindStringSubmatch(line)
		if len(match) == 0 {
			t.Errorf("unrecognized step in line %d: %s", i+1, line)
			errors++
			continue
		}
		switch match[1] {
		case "ak", "http", "wait":
			if !actions.MatchString(line) {
				t.Errorf("invalid action in line %d: %s", i+1, line)
				errors++
			}
		case "output", "return":
			if !akCheckOutput.MatchString(line) && !akCheckReturn.MatchString(line) {
				t.Errorf("invalid AK check in line %d: %s", i+1, line)
				errors++
			}
		case "req":
			if !httpCustomHeader.MatchString(line) && !httpCustomBody.MatchString(line) {
				t.Errorf("invalid HTTP request customization in line %d: %s", i+1, line)
				errors++
			}
		case "resp":
			if !httpCheckOutput.MatchString(line) && !httpCheckStatus.MatchString(line) && !httpCheckHeader.MatchString(line) {
				t.Errorf("invalid HTTP response check in line %d: %s", i+1, line)
				errors++
			}
		}
	}

	// We use many "t.Errorf" and an eventual "t.Fatalf" to report as many errors
	// as possible, instead of many "t.Fatalf" to report only the first error.
	if errors > 0 {
		t.Fatalf("found %d test script errors", errors)
	}

	return lines
}
