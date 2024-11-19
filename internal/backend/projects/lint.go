/*
	Project linting

We don't want to run a Python/NodeJS process on every call to lint, so we're using regular expression.
This means we won't be right every time but close enought for now.
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
	"strings"

	"go.autokitteh.dev/autokitteh/internal/backend/projectsgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/manifest"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Checker func(projectID sdktypes.ProjectID, manifest *manifest.Manifest, resources map[string][]byte) []*sdktypes.CheckViolation

var (
	lintCheckers []Checker
)

const manifestFile = "autokitteh.yaml"

func init() {
	lintCheckers = []Checker{
		// Generic
		checkNoTriggers,
		checkEmptyVars,
		checkSize,
		checkConnectionNames,
		checkTriggerNames,

		// Runtime
		checkHandlers,
		checkCodeConnections,
	}
}

func Validate(projectID sdktypes.ProjectID, manifest *manifest.Manifest, resources map[string][]byte) []*sdktypes.CheckViolation {
	var vs []*sdktypes.CheckViolation
	for _, checker := range lintCheckers {
		vs = append(vs, checker(projectID, manifest, resources)...)
	}

	return vs
}

var Rules = map[string]string{ // ID -> Descrption
	"E1": "No triggers defined",
	"E2": "Project size too large",
	"E3": "Duplicate connection name",
	"E4": "Duplicate trigger name",
	"E5": "Bad `call` format",
	"E6": "File not found",
	"E7": "Syntax error",
	"E8": "Missing handler",
	"E9": "Nonexisting connection",

	"W1": "Empty variable",
}

func checkNoTriggers(_ sdktypes.ProjectID, m *manifest.Manifest, _ map[string][]byte) []*sdktypes.CheckViolation {
	if len(m.Project.Triggers) == 0 {
		return []*sdktypes.CheckViolation{
			{
				FileName: manifestFile,
				Level:    sdktypes.ViolationError,
				Message:  "no triggers",
				RuleId:   "E1",
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
				FileName: manifestFile,
				Level:    sdktypes.ViolationWarning,
				Message:  fmt.Sprintf("%q is empty", v.Name),
				RuleId:   "W1",
			})
		}
	}

	for _, conn := range m.Project.Connections {
		for _, v := range conn.Vars {
			if v.Value == "" {
				vs = append(vs, &sdktypes.CheckViolation{
					FileName: manifestFile,
					Level:    sdktypes.ViolationWarning,
					Message:  fmt.Sprintf("connection %q: %q is empty", conn.Name, v.Name),
					RuleId:   "W1",
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
				FileName: manifestFile,
				Level:    sdktypes.ViolationError,
				Message:  fmt.Sprintf("project size (%.2fMB) exceeds limt of %dMB", sizeMB, maxProjectSize/mb),
				RuleId:   "E2",
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
		if count > 1 {
			vs = append(vs, &sdktypes.CheckViolation{
				FileName: manifestFile,
				Level:    sdktypes.ViolationWarning,
				Message:  fmt.Sprintf("%d connections are named %q", count, name),
				RuleId:   "E3",
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
		if count > 1 {
			vs = append(vs, &sdktypes.CheckViolation{
				FileName: manifestFile,
				Level:    sdktypes.ViolationWarning,
				Message:  fmt.Sprintf("%d triggers are named %q", count, name),
				RuleId:   "E4",
			})
		}
	}

	return vs
}

func checkHandlers(_ sdktypes.ProjectID, m *manifest.Manifest, resources map[string][]byte) []*sdktypes.CheckViolation {
	var vs []*sdktypes.CheckViolation

	for _, t := range m.Project.Triggers {
		fileName, handler, found := strings.Cut(t.Call, ":")
		if !found {
			vs = append(vs, &sdktypes.CheckViolation{
				FileName: manifestFile,
				Level:    sdktypes.ViolationError,
				Message:  fmt.Sprintf(`%q - bad call definition (should be something like "handler.py:on_event"`, t.Call),
				RuleId:   "E5",
			})
			continue
		}

		data, ok := resources[fileName]
		if !ok {
			vs = append(vs, &sdktypes.CheckViolation{
				FileName: manifestFile,
				Level:    sdktypes.ViolationError,
				Message:  fmt.Sprintf("file %q not found", fileName),
				RuleId:   "E6",
			})
			continue
		}

		exports, err := fileExports(fileName, data)
		if err != nil {
			vs = append(vs, &sdktypes.CheckViolation{
				FileName: manifestFile,
				Level:    sdktypes.ViolationError,
				Message:  fmt.Sprintf("can't parse %q - %s", fileName, err),
				RuleId:   "E7",
			})
			continue
		}

		if !slices.Contains(exports, handler) {
			vs = append(vs, &sdktypes.CheckViolation{
				FileName: manifestFile,
				Level:    sdktypes.ViolationError,
				Message:  fmt.Sprintf("%q not found in %q", handler, fileName),
				RuleId:   "E8",
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
					FileName: fileName,
					Line:     conn.Line,
					Message:  fmt.Sprintf("%q - non existing connection", conn.Name),
					RuleId:   "E9",
				})

			}
		}
	}

	return vs
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
	"redis_client":           true,
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
		// TODO: Do we want to return empty slice without error?
		return nil, fmt.Errorf("no runtime for %q", fileName)
	}

	names, err := exportsFn(data)
	if err != nil {
		return nil, err
	}

	return names, nil
}
