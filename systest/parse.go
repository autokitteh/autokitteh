package systest

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"golang.org/x/tools/txtar"
)

const (
	tempFilePerm = 0o644 // -rw-r--r--
)

var (
	steps = regexp.MustCompile(`^(ak|http|output|req|resp|return)\s`)

	// ak *
	// http <get|post> *
	actions = regexp.MustCompile(`^(ak|http\s+(get|post))\s+(.+)`)

	// output <equals|contains|regex> [file] *
	akCheckOutput = regexp.MustCompile(`^output\s+(equals|contains|regex)\s+(file\s+)?(.+)`)
	// return code == <int>
	akCheckReturn = regexp.MustCompile(`^return\s+code\s*==\s*(\d+)$`)

	// req header <name> = <value>
	httpCustom1 = regexp.MustCompile(`^req\s+header\s+([\w-]+)\s*=\s*(.+)`)
	// req body [file] *
	httpCustom2 = regexp.MustCompile(`^req\s+body\s+(file\s+)?(.+)`)

	// resp <body|redirect> <equals|contains|regex> [file] *
	httpCheck1 = regexp.MustCompile(`^resp\s+(body|redirect)\s+(equals|contains|regex)\s+(file\s+)?(.+)`)
	// resp code == <int>
	httpCheck2 = regexp.MustCompile(`^resp\s+code\s*==\s*(\d+)$`)
	// resp header <name> == <value>
	httpCheck3 = regexp.MustCompile(`^resp\s+header\s+([\w-]+)\s*==\s*(.+)`)
)

func readTestFile(t *testing.T, path string) []string {
	a, err := txtar.ParseFile(path)
	if err != nil {
		t.Fatalf("failed to load file: %v", err)
	}

	if len(a.Comment) == 0 {
		t.Fatalf("nothing to do in %q: txtar comment section is empty", path)
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
}

func writeEmbeddedFiles(t *testing.T, fs []txtar.File) {
	for _, f := range fs {
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
		case "ak", "http":
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
			if !httpCustom1.MatchString(line) && !httpCustom2.MatchString(line) {
				t.Errorf("invalid HTTP request customization in line %d: %s", i+1, line)
				errors++
			}
		case "resp":
			if !httpCheck1.MatchString(line) && !httpCheck2.MatchString(line) && !httpCheck3.MatchString(line) {
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
