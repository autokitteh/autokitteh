package projects

import (
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/manifest"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	lintChecks []func(manifest *manifest.Manifest) []*sdktypes.CheckViolation
)

const manifestFile = "autokitteh.yaml"

func init() {
	lintChecks = []func(manifest *manifest.Manifest) []*sdktypes.CheckViolation{
		checkNoTriggers,
		checkEmptyVars,
	}
}

func checkNoTriggers(m *manifest.Manifest) []*sdktypes.CheckViolation {
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

func checkEmptyVars(m *manifest.Manifest) []*sdktypes.CheckViolation {
	var violations []*sdktypes.CheckViolation
	for _, v := range m.Project.Vars {
		if v.Value == "" {
			violations = append(violations, &sdktypes.CheckViolation{
				FileName: manifestFile,
				Level:    sdktypes.ViolationWarning,
				Message:  fmt.Sprintf("%q is empty", v.Name),
			})
		}
	}

	return violations
}
