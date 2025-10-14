/*
	Project linting

We don't want to run a Python/NodeJS process on every call to lint, so we're using regular expression.
This means we won't be right every time but close enough for now.
Later we can think of running a lint/lsp server for these calls.

WARNING: manifest.Manifest contains pointers (such as Project), check for nils.
*/
package projects

import (
	"bufio"
	"bytes"
	"errors"
	"path"
	"regexp"
	"slices"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/backend/integrations"
	"go.autokitteh.dev/autokitteh/internal/backend/projectsgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/manifest"
	pysdk "go.autokitteh.dev/autokitteh/runtimes/pythonrt/py-sdk"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type checker func(projectID sdktypes.ProjectID, manifest *manifest.Manifest, resources map[string][]byte) []*sdktypes.CheckViolation

var lintCheckers = []checker{
	// Generic
	checkConnectionsNames,
	checkIntegrationsNames,
	checkEmptyVars,
	checkProjectName,
	checkSize,
	checkNoTriggers,
	checkTriggersNames,

	// Runtime
	checkCodeConnections,
	checkHandlers,
	checkPyRequirements,
}

const manifestFilePath = "autokitteh.yaml"

var pythonRequirementPackageNameRegexp = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_.-]+)`)

func Validate(projectID sdktypes.ProjectID, manifestData []byte, resources map[string][]byte) []*sdktypes.CheckViolation {
	manifest, err := manifest.Read(manifestData)
	if err != nil {
		return []*sdktypes.CheckViolation{
			sdktypes.NewCheckViolationf(
				manifestFilePath,
				sdktypes.InvalidManifestRuleID,
				"bad manifest - %s",
				err,
			),
		}
	}

	var vs []*sdktypes.CheckViolation
	for _, checker := range lintCheckers {
		vs = append(vs, checker(projectID, manifest, resources)...)
	}

	return vs
}

func checkNoTriggers(_ sdktypes.ProjectID, m *manifest.Manifest, _ map[string][]byte) []*sdktypes.CheckViolation {
	if m.Project == nil || len(m.Project.Triggers) == 0 {
		return []*sdktypes.CheckViolation{
			sdktypes.NewCheckViolationf(
				manifestFilePath,
				sdktypes.NoTriggersDefinedRuleID,
				"no triggers defined",
			),
		}
	}

	return nil
}

func checkEmptyVars(_ sdktypes.ProjectID, m *manifest.Manifest, _ map[string][]byte) []*sdktypes.CheckViolation {
	if m.Project == nil || len(m.Project.Vars) == 0 {
		return nil
	}

	var vs []*sdktypes.CheckViolation
	for _, v := range m.Project.Vars {
		if v.Value == "" {
			vs = append(vs, sdktypes.NewCheckViolationf(
				manifestFilePath,
				sdktypes.EmptyVariableRuleID,
				"project variable %q is empty",
				v.Name,
			))
		}
	}

	for _, conn := range m.Project.Connections {
		for _, v := range conn.Vars {
			if v.Value == "" {
				vs = append(vs, sdktypes.NewCheckViolationf(
					manifestFilePath,
					sdktypes.EmptyVariableRuleID,
					"connection %q variable %q is empty",
					conn.Name,
					v.Name,
				))
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
			sdktypes.NewCheckViolationf(
				manifestFilePath,
				sdktypes.ProjectSizeTooLargeRuleID,
				"project size (%.2fMB) exceeds limit of %dMB",
				sizeMB,
				maxProjectSize/mb,
			),
		}
	}

	return nil
}

func checkConnectionsNames(_ sdktypes.ProjectID, m *manifest.Manifest, _ map[string][]byte) []*sdktypes.CheckViolation {
	if m.Project == nil || len(m.Project.Connections) == 0 {
		return nil
	}

	names := make(map[string]int) // name -> count
	for _, c := range m.Project.Connections {
		names[c.Name]++
	}

	var vs []*sdktypes.CheckViolation
	for name, count := range names {
		if _, err := sdktypes.ParseSymbol(name); err != nil {
			vs = append(vs, sdktypes.NewCheckViolationf(
				manifestFilePath,
				sdktypes.MalformedNameRuleID,
				"%q - malformed name (%s)",
				name,
				err,
			))
		}

		if count > 1 {
			vs = append(vs, sdktypes.NewCheckViolationf(
				manifestFilePath,
				sdktypes.DuplicateConnectionNameRuleID,
				"%d connections are named %q",
				count,
				name,
			))
		}
	}

	return vs
}

func checkIntegrationsNames(_ sdktypes.ProjectID, m *manifest.Manifest, _ map[string][]byte) (vs []*sdktypes.CheckViolation) {
	if m.Project == nil || len(m.Project.Connections) == 0 {
		return nil
	}

	hasIntegration := kittehs.ContainedIn(integrations.Names()...)

	for _, c := range m.Project.Connections {
		name := c.IntegrationKey

		if _, err := sdktypes.ParseSymbol(name); err != nil {
			vs = append(
				vs,
				sdktypes.NewCheckViolationf(
					manifestFilePath,
					sdktypes.MalformedNameRuleID,
					"%q - malformed name (%s)",
					name,
					err,
				))
		}

		if !hasIntegration(name) {
			vs = append(
				vs,
				sdktypes.NewCheckViolationf(
					manifestFilePath,
					sdktypes.UnknownIntegrationRuleID,
					"%q - unknown integration",
					name,
				))
		}
	}

	return vs
}

func checkTriggersNames(_ sdktypes.ProjectID, m *manifest.Manifest, _ map[string][]byte) []*sdktypes.CheckViolation {
	if m.Project == nil || len(m.Project.Triggers) == 0 {
		return nil
	}

	names := make(map[string]int) // name -> count
	for _, c := range m.Project.Triggers {
		names[c.Name]++
	}

	var vs []*sdktypes.CheckViolation
	for name, count := range names {
		if _, err := sdktypes.ParseSymbol(name); err != nil {
			vs = append(vs, sdktypes.NewCheckViolationf(
				manifestFilePath,
				sdktypes.MalformedNameRuleID,
				"%q - malformed name (%s)",
				name,
				err,
			))
		}

		if count > 1 {
			vs = append(vs, sdktypes.NewCheckViolationf(
				manifestFilePath,
				sdktypes.DuplicateTriggerNameRuleID,
				"%d triggers are named %q",
				count,
				name,
			))
		}
	}

	return vs
}

func checkHandlers(_ sdktypes.ProjectID, m *manifest.Manifest, resources map[string][]byte) []*sdktypes.CheckViolation {
	if m.Project == nil || len(m.Project.Triggers) == 0 {
		return nil
	}

	var vs []*sdktypes.CheckViolation

	for _, t := range m.Project.Triggers {
		if err := sdktypes.ValidateEventFilterField(t.Filter); err != nil {
			vs = append(vs, sdktypes.NewCheckViolationf(
				manifestFilePath,
				sdktypes.InvalidEventFilterRuleID,
				"invalid event filter %q - %s",
				t.Filter,
				err,
			))

			continue
		}

		// It OK to have a trigger without "Call"
		if t.Call == "" {
			continue
		}

		loc, err := sdktypes.StrictParseCodeLocation(t.Call)
		if err != nil {
			vs = append(vs, sdktypes.NewCheckViolationf(
				manifestFilePath,
				sdktypes.BadCallFormatRuleID,
				`%q - bad call definition (should be something like "handler.py:on_event")`,
				t.Call,
			))

			continue
		}

		fileName := loc.Path()
		data, ok := resources[fileName]
		if !ok {
			vs = append(vs, sdktypes.NewCheckViolationf(
				manifestFilePath,
				sdktypes.FileNotFoundRuleID,
				"file %q not found",
				fileName,
			))
			continue
		}

		exports, err := fileExports(fileName, data)
		if err != nil {
			if errors.Is(err, sdkerrors.ErrNotImplemented) {
				vs = append(vs, sdktypes.NewCheckViolationf(
					manifestFilePath,
					sdktypes.FileCannotExportRuleID,
					"file %q cannot export",
					fileName,
				))
				continue
			}

			vs = append(vs, sdktypes.NewCheckViolationf(
				manifestFilePath,
				sdktypes.SyntaxErrorRuleID,
				"can't parse %q - %s",
				fileName,
				err,
			))
			continue
		}

		handler := loc.Name()
		if !slices.Contains(exports, handler) {
			vs = append(vs, sdktypes.NewCheckViolationf(
				manifestFilePath,
				sdktypes.MissingHandlerRuleID,
				"%q not found in %q",
				handler,
				fileName,
			))
			continue
		}
	}

	return vs
}

func checkCodeConnections(_ sdktypes.ProjectID, m *manifest.Manifest, resources map[string][]byte) []*sdktypes.CheckViolation {
	defConns := make(map[string]bool)
	if m.Project != nil {
		for _, conn := range m.Project.Connections {
			defConns[conn.Name] = true
			// TODO: Should we use GetKey as well?
		}
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
				vs = append(vs, sdktypes.NewCheckViolationf(
					manifestFilePath,
					sdktypes.NonexistingConnectionRuleID,
					"%q - non existing connection",
					conn.Name,
				))
			}
		}
	}

	return vs
}

func checkProjectName(_ sdktypes.ProjectID, m *manifest.Manifest, resources map[string][]byte) []*sdktypes.CheckViolation {
	if m.Project == nil || m.Project.Name == "" {
		return []*sdktypes.CheckViolation{
			sdktypes.NewCheckViolationf(
				manifestFilePath,
				sdktypes.InvalidManifestRuleID,
				"bad project name",
			),
		}
	}

	if _, err := sdktypes.ParseSymbol(m.Project.Name); err != nil {
		return []*sdktypes.CheckViolation{
			sdktypes.NewCheckViolationf(
				manifestFilePath,
				sdktypes.MalformedNameRuleID,
				"%q - bad project name (%s)",
				m.Project.Name,
				err,
			),
		}
	}

	return nil
}

func checkPyRequirements(_ sdktypes.ProjectID, _ *manifest.Manifest, resources map[string][]byte) (vs []*sdktypes.CheckViolation) {
	const path = "requirements.txt"

	txt, ok := resources[path]
	if !ok {
		// No requirements.txt, nothing to check
		return nil
	}

	inEmbeddedDependency := kittehs.ContainedIn(pysdk.Dependencies()...)

	// parse requirements.txt line by line
	lines := strings.Split(string(txt), "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		// parse line using regex to get the first part (package name)
		matches := pythonRequirementPackageNameRegexp.FindStringSubmatch(line)
		if matches == nil {
			vs = append(
				vs,
				sdktypes.NewCheckViolationf(
					path,
					sdktypes.InvalidPyRequirementsRuleID,
					"invalid requirement - line %d is not a valid requirement",
					i+1,
				),
			)
			continue
		}

		name := matches[1]

		if inEmbeddedDependency(name) {
			vs = append(
				vs,
				sdktypes.NewCheckViolationf(
					path,
					sdktypes.PyRequirementsPackageAlreadyInstalledRuleID,
					"dependency %q is already included in the Autokitteh Python runtime",
					name,
				),
			)
		}
	}

	return
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
var pyClientFns = kittehs.ListToBoolSet(pysdk.ClientNames())

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
		return nil, sdkerrors.ErrNotImplemented
	}

	names, err := exportsFn(data)
	if err != nil {
		return nil, err
	}

	return names, nil
}
