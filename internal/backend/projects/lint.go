package projects

import (
	"context"
	"fmt"
	"os"
	"path"
	"slices"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/backend/projectsgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/manifest"
	"go.autokitteh.dev/autokitteh/runtimes/pythonrt"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.uber.org/zap"
)

type Checker func(projectID sdktypes.ProjectID, manifest *manifest.Manifest, resources map[string][]byte) []*sdktypes.CheckViolation

var (
	lintCheckers []Checker
)

const manifestFile = "autokitteh.yaml"

func init() {
	lintCheckers = []Checker{
		checkNoTriggers,
		checkEmptyVars,
	}
}

func Validate(projectID sdktypes.ProjectID, manifest *manifest.Manifest, resources map[string][]byte) []*sdktypes.CheckViolation {
	var vs []*sdktypes.CheckViolation
	for _, checker := range lintCheckers {
		vs = append(vs, checker(projectID, manifest, resources)...)
	}

	return vs
}

func checkNoTriggers(_ sdktypes.ProjectID, m *manifest.Manifest, _ map[string][]byte) []*sdktypes.CheckViolation {
	if len(m.Project.Triggers) == 0 {
		return []*sdktypes.CheckViolation{
			{
				FileName: manifestFile,
				Level:    sdktypes.ViolationError,
				Message:  "no triggers",
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
			})
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
			})
			continue
		}

		data, ok := resources[fileName]
		if !ok {
			vs = append(vs, &sdktypes.CheckViolation{
				FileName: manifestFile,
				Level:    sdktypes.ViolationError,
				Message:  fmt.Sprintf("file %q not found", fileName),
			})
			continue
		}

		exports, err := fileExports(fileName, data)
		if err != nil {
			vs = append(vs, &sdktypes.CheckViolation{
				FileName: manifestFile,
				Level:    sdktypes.ViolationError,
				Message:  fmt.Sprintf("can't parse %q - %s", fileName, err),
			})
			continue
		}

		if !slices.Contains(exports, handler) {
			vs = append(vs, &sdktypes.CheckViolation{
				FileName: manifestFile,
				Level:    sdktypes.ViolationError,
				Message:  fmt.Sprintf("%q not found in %q", handler, fileName),
			})
			continue
		}
	}

	return vs
}

func pyExports(fileName string) ([]string, error) {
	cfg := pythonrt.Config{
		RunnerType: "local",
	}
	log := zap.NewExample()
	getLocalAddr := func() string { return "127.0.0.1" }

	rt, err := pythonrt.New(&cfg, log, getLocalAddr)
	if err != nil {
		return nil, err
	}

	runner, err := rt.New()
	if err != nil {
		return nil, err
	}

	cbs := sdkservices.RunCallbacks{
		Load: func(ctx context.Context, rid sdktypes.RunID, path string) (map[string]sdktypes.Value, error) {
			return nil, nil
		},
	}

	run, err := runner.Run(context.Background(), sdktypes.NewRunID(), sdktypes.NewSessionID(), fileName, nil, nil, &cbs)
	if err != nil {
		return nil, err
	}

	names := kittehs.TransformMapToList(run.Values(), func(k string, v sdktypes.Value) string { return k })
	return names, nil
}

var exportsByExt = map[string]func(string) ([]string, error){
	".py": pyExports,
}

func fileExports(fileName string, data []byte) ([]string, error) {
	ext := path.Ext(fileName)
	exportsFn, ok := exportsByExt[ext]
	if !ok {
		return nil, fmt.Errorf("no runtime for %q", fileName)
	}

	tmp, err := os.CreateTemp("", path.Ext(fileName))
	if err != nil {
		return nil, err
	}
	defer tmp.Close()

	if _, err := tmp.Write(data); err != nil {
		return nil, err
	}
	tmp.Close()

	names, err := exportsFn(fileName)
	if err != nil {
		return nil, err
	}

	return names, nil
}
