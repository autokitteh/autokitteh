package projects

import (
	"fmt"

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
	mb             = 1 << 20
	maxProjectSize = 10 * mb
)

func checkSize(_ sdktypes.ProjectID, _ *manifest.Manifest, resources map[string][]byte) []*sdktypes.CheckViolation {
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
