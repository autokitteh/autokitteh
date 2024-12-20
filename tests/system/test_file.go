package systest

import (
	"embed"
	"encoding/json"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"golang.org/x/tools/txtar"
	"gopkg.in/yaml.v2"
)

//go:embed testdata/*
var testDataFS embed.FS

type testFile struct {
	steps  []string
	config testConfig
	a      *txtar.Archive
}

func readTestFile(t *testing.T, path string) (*testFile, error) {
	bs, err := fs.ReadFile(testDataFS, path)
	if err != nil {
		t.Fatalf("failed to load file: %v", err)
	}

	a := txtar.Parse(bs)

	if len(a.Comment) == 0 {
		t.Fatalf("nothing to do in %q: txtar comment section is empty", path)
	}

	for i, f := range a.Files {
		// Support embedded txtars.
		if filepath.Ext(f.Name) == ".txtar" {
			a.Files[i].Data = []byte(strings.ReplaceAll(string(f.Data), "~~", "--"))
		}
	}

	return parseTestFile(t, a), nil
}

func prepTestFiles(t *testing.T, a *txtar.Archive) *testFile {
	useTempDir(t)
	writeEmbeddedFiles(t, a.Files)
	return parseTestFile(t, a)
}

func parseTestFile(t *testing.T, a *txtar.Archive) *testFile {
	var cfg testConfig

	for _, f := range a.Files {
		if f.Name == "test-config.json" {
			if err := json.Unmarshal(f.Data, &cfg); err != nil {
				t.Fatalf("failed to parse server config: %v", err)
			}
		} else if f.Name == "test-config.yaml" {
			if err := yaml.Unmarshal(f.Data, &cfg); err != nil {
				t.Fatalf("failed to parse server config: %v", err)
			}
		}
	}

	lines := strings.Split(string(a.Comment), "\n")
	errors := 0
	for i, line := range lines {
		// Trim redundant whitespaces and single-line comments
		// (but don't discard empty lines, to preserve line numbers).
		lines[i] = strings.TrimSpace(line)

		lines[i] = regexp.MustCompile(`^\s*#.*`).ReplaceAllString(line, "")

		lines[i] = expandConsts(lines[i])

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
		case "exit":
			// nop
		case "ak", "http", "wait", "setenv", "user":
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
		case "capture_jq":
			if !jqCheck.MatchString(line) {
				t.Errorf("invalid jq capture in line %d: %s", i+1, line)
				errors++
			}
		}
	}

	// We use many "t.Errorf" and an eventual "t.Fatalf" to report as many errors
	// as possible, instead of many "t.Fatalf" to report only the first error.
	if errors > 0 {
		t.Fatalf("found %d test script errors", errors)
	}

	return &testFile{lines, cfg, a}
}
