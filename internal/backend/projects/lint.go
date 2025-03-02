/*
	Project linting

We don't want to run a Python/NodeJS process on every call to lint, so we're using regular expression.
This means we won't be right every time but close enough for now.
Later we can think of running a lint/lsp server for these calls.
*/
package projects

import (
	"bufio"
	"bytes"
	"fmt"
	"path"
	"regexp"
	"slices"

	"go.autokitteh.dev/autokitteh/internal/backend/projectsgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/manifest"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Checker func(projectID sdktypes.ProjectID, manifest *manifest.Manifest, resources map[string][]byte) []*sdktypes.CheckViolation

var lintCheckers []Checker

const manifestFilePath = "autokitteh.yaml"

func init() {
	// Please keep the groups sorted alphabetically
	lintCheckers = []Checker{
		// Generic
		checkConnectionNames,
		checkEmptyVars,
		checkNoTriggers,
		checkProjectName,
		checkSize,
		checkTriggerNames,

		// Runtime
		checkCodeConnections,
		checkHandlers,
	}
}

func Validate(projectID sdktypes.ProjectID, manifest *manifest.Manifest, resources map[string][]byte) []*sdktypes.CheckViolation {
	var vs []*sdktypes.CheckViolation
	for _, checker := range lintCheckers {
		vs = append(vs, checker(projectID, manifest, resources)...)
	}

	return vs
}

var Rules = map[string]string{ // ID -> Description
	"E1": "Project size too large",
	"E2": "Duplicate connection name",
	"E3": "Duplicate trigger name",
	"E4": "Bad `call` format",
	"E5": "File not found",
	"E6": "Syntax error",
	"E7": "Missing handler",
	"E8": "Nonexisting connection",
	"E9": "Malformed name",

	"W1": "Empty variable",
	"W2": "No triggers defined",
}

func checkNoTriggers(_ sdktypes.ProjectID, m *manifest.Manifest, _ map[string][]byte) []*sdktypes.CheckViolation {
	if len(m.Project.Triggers) == 0 {
		return []*sdktypes.CheckViolation{
			{
				Location: &sdktypes.CodeLocationPB{
					Path: manifestFilePath,
				},
				Level:   sdktypes.ViolationWarning,
				Message: "no triggers",
				RuleId:  "W2",
			},
		}
	}

	return nil
}

func checkEmptyVars(_ sdktypes.ProjectID, m *manifest.Manifest, _ map[string][]byte) []*sdktypes.CheckViolation {
	var vs []*sdktypes.CheckViolation
	for _, v := range m.Project.Vars {
		if v.Value == "" {
			vs = append(vs, &sdktypes.CheckViolation{
				Location: &sdktypes.CodeLocationPB{
					Path: manifestFilePath,
				},
				Level:   sdktypes.ViolationWarning,
				Message: fmt.Sprintf("variable %q is empty", v.Name),
				RuleId:  "W1",
			})
		}
	}

	for _, conn := range m.Project.Connections {
		for _, v := range conn.Vars {
			if v.Value == "" {
				vs = append(vs, &sdktypes.CheckViolation{
					Location: &sdktypes.CodeLocationPB{
						Path: manifestFilePath,
					},
					Level:   sdktypes.ViolationWarning,
					Message: fmt.Sprintf("connection %q variable %q is empty", conn.Name, v.Name),
					RuleId:  "W1",
				})
			}
		}
	}

	return vs
}

const (
	mb = 1 << 20
)

func checkSize(_ sdktypes.ProjectID, _ *manifest.Manifest, resources map[string][]byte) []*sdktypes.CheckViolation {
	maxProjectSize := projectsgrpcsvc.Configs.Default.MaxUploadSize

	total := 0
	for _, data := range resources {
		total += len(data)
	}

	if total > maxProjectSize {
		sizeMB := float64(total) / mb
		return []*sdktypes.CheckViolation{
			{
				Location: &sdktypes.CodeLocationPB{
					Path: manifestFilePath,
				},
				Level:   sdktypes.ViolationError,
				Message: fmt.Sprintf("project size (%.2fMB) exceeds limt of %dMB", sizeMB, maxProjectSize/mb),
				RuleId:  "E2",
			},
		}
	}

	return nil
}

func checkConnectionNames(_ sdktypes.ProjectID, m *manifest.Manifest, _ map[string][]byte) []*sdktypes.CheckViolation {
	names := make(map[string]int) // name -> count
	for _, c := range m.Project.Connections {
		names[c.Name]++
	}

	var vs []*sdktypes.CheckViolation
	for name, count := range names {
		if _, err := sdktypes.ParseSymbol(name); err != nil {
			vs = append(vs, &sdktypes.CheckViolation{
				Location: &sdktypes.CodeLocationPB{
					Path: manifestFilePath,
				},
				Level:   sdktypes.ViolationError,
				Message: fmt.Sprintf("%q - malformed name (%s)", name, err),
				RuleId:  "E10",
			})
		}

		if count > 1 {
			vs = append(vs, &sdktypes.CheckViolation{
				Location: &sdktypes.CodeLocationPB{
					Path: manifestFilePath,
				},
				Level:   sdktypes.ViolationWarning,
				Message: fmt.Sprintf("%d connections are named %q", count, name),
				RuleId:  "E3",
			})
		}
	}

	return vs
}

func checkTriggerNames(_ sdktypes.ProjectID, m *manifest.Manifest, _ map[string][]byte) []*sdktypes.CheckViolation {
	names := make(map[string]int) // name -> count
	for _, c := range m.Project.Triggers {
		names[c.Name]++
	}

	var vs []*sdktypes.CheckViolation
	for name, count := range names {
		if _, err := sdktypes.ParseSymbol(name); err != nil {
			vs = append(vs, &sdktypes.CheckViolation{
				Location: &sdktypes.CodeLocationPB{
					Path: manifestFilePath,
				},
				Level:   sdktypes.ViolationError,
				Message: fmt.Sprintf("%q - malformed name (%s)", name, err),
				RuleId:  "E10",
			})
		}

		if count > 1 {
			vs = append(vs, &sdktypes.CheckViolation{
				Location: &sdktypes.CodeLocationPB{
					Path: manifestFilePath,
				},
				Level:   sdktypes.ViolationWarning,
				Message: fmt.Sprintf("%d triggers are named %q", count, name),
				RuleId:  "E4",
			})
		}
	}

	return vs
}

// parseCall parses a call string like "handler.py:on_event" and return file name and function name.

func checkHandlers(_ sdktypes.ProjectID, m *manifest.Manifest, resources map[string][]byte) []*sdktypes.CheckViolation {
	var vs []*sdktypes.CheckViolation

	for _, t := range m.Project.Triggers {
		// It OK to have a trigger without "Call"
		if t.Call == "" {
			continue
		}

		loc, err := sdktypes.StrictParseCodeLocation(t.Call)
		if err != nil {
			vs = append(vs, &sdktypes.CheckViolation{
				Location: &sdktypes.CodeLocationPB{
					Path: manifestFilePath,
				},
				Level:   sdktypes.ViolationError,
				Message: fmt.Sprintf(`%q - bad call definition (should be something like "handler.py:on_event")`, t.Call),
				RuleId:  "E5",
			})
			continue
		}

		fileName := loc.Path()
		data, ok := resources[fileName]
		if !ok {
			vs = append(vs, &sdktypes.CheckViolation{
				Location: &sdktypes.CodeLocationPB{
					Path: manifestFilePath,
				},
				Level:   sdktypes.ViolationError,
				Message: fmt.Sprintf("file %q not found", fileName),
				RuleId:  "E6",
			})
			continue
		}

		exports, err := fileExports(fileName, data)
		if err != nil {
			vs = append(vs, &sdktypes.CheckViolation{
				Location: &sdktypes.CodeLocationPB{
					Path: manifestFilePath,
				},
				Level:   sdktypes.ViolationError,
				Message: fmt.Sprintf("can't parse %q - %s", fileName, err),
				RuleId:  "E7",
			})
			continue
		}

		handler := loc.Name()
		if exports != nil && !slices.Contains(exports, handler) {
			vs = append(vs, &sdktypes.CheckViolation{
				Location: &sdktypes.CodeLocationPB{
					Path: manifestFilePath,
				},
				Level:   sdktypes.ViolationError,
				Message: fmt.Sprintf("%q not found in %q", handler, fileName),
				RuleId:  "E8",
			})
			continue
		}
	}

	return vs
}

func checkCodeConnections(_ sdktypes.ProjectID, m *manifest.Manifest, resources map[string][]byte) []*sdktypes.CheckViolation {
	defConns := make(map[string]bool)
	for _, conn := range m.Project.Connections {
		defConns[conn.Name] = true
		// TODO: Should we use GetKey as well?
	}

	var vs []*sdktypes.CheckViolation
	for fileName, code := range resources {
		fn := codeConnFns[path.Ext(fileName)]
		if fn == nil {
			// TODO: Log? Error?
			continue
		}

		for _, conn := range fn(code) {
			if !defConns[conn.Name] {
				vs = append(vs, &sdktypes.CheckViolation{
					Location: &sdktypes.CodeLocationPB{
						Path: manifestFilePath,
						Row:  conn.Line,
					},
					Message: fmt.Sprintf("%q - non existing connection", conn.Name),
					Level:   sdktypes.ViolationError,
					RuleId:  "E9",
				})
			}
		}
	}

	return vs
}

func checkProjectName(_ sdktypes.ProjectID, m *manifest.Manifest, resources map[string][]byte) []*sdktypes.CheckViolation {
	if _, err := sdktypes.ParseSymbol(m.Project.Name); err != nil {
		return []*sdktypes.CheckViolation{
			{
				Location: &sdktypes.CodeLocationPB{
					Path: manifestFilePath,
				},
				Message: fmt.Sprintf("%q - bad project name (%s)", m.Project.Name, err),
				Level:   sdktypes.ViolationError,
				RuleId:  "E10",
			},
		}
	}

	return nil
}

type connCall struct {
	Name string
	Line uint32
}

var codeConnFns = map[string]func([]byte) []connCall{
	".py": pyConnCalls,
}

// asana_client("asana1") -> asana1
// boto3_client(conn_name) -> boto3_client
var pyConnCallRe = regexp.MustCompile(`(\w+)\s*\(('|")?([A-Za-z_]\w*)('|")?`)

func pyConnCalls(data []byte) []connCall {
	var calls []connCall
	var lnum uint32 = 0

	s := bufio.NewScanner(bytes.NewReader(data))
	for s.Scan() {
		lnum++
		match := pyConnCallRe.FindStringSubmatch(s.Text())
		if match == nil {
			continue
		}

		// FIXME: Ignore comments (e.g. "# boto3_client('aws1')")

		fnName, conn := match[1], match[3]
		if !pyClientFns[fnName] {
			continue
		}

		calls = append(calls, connCall{conn, lnum})
	}

	if err := s.Err(); err != nil {
		// TODO: Log
		return nil
	}

	return calls
}

// Should be in sync with runtimes/pythonrt/py-sdk/autokitteh
var pyClientFns = map[string]bool{
	"asana_client":           true,
	"boto3_client":           true,
	"confluence_client":      true,
	"discord_client":         true,
	"github_client":          true,
	"gmail_client":           true,
	"google_calendar_client": true,
	"google_drive_client":    true,
	"google_forms_client":    true,
	"google_sheets_client":   true,
	"jira_client":            true,
	"openai_client":          true,
	"slack_client":           true,
	"twilio_client":          true,
}

var callRe = regexp.MustCompile(`^(def|class)\s+(\w+)\(`)

func pyExports(data []byte) ([]string, error) {
	var names []string

	s := bufio.NewScanner(bytes.NewReader(data))

	for s.Scan() {
		matches := callRe.FindStringSubmatch(s.Text())
		if matches == nil {
			continue
		}
		names = append(names, matches[2])
	}

	if err := s.Err(); err != nil {
		return nil, err
	}

	return names, nil
}

var exportsByExt = map[string]func([]byte) ([]string, error){
	".py": pyExports,
}

func fileExports(fileName string, data []byte) ([]string, error) {
	ext := path.Ext(fileName)
	exportsFn, ok := exportsByExt[ext]
	if !ok {
		return nil, nil
	}

	names, err := exportsFn(data)
	if err != nil {
		return nil, err
	}

	return names, nil
}
