package manifest

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v2"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

func Read(data []byte, path string) (*Manifest, error) {
	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("invalid YAML input: %w", err)
	}

	if manifest.Project != nil && manifest.Project.RootURL == "" {
		manifest.Project.RootURL = filepath.Dir(path)
	}

	res, err := gojsonschema.Validate(
		gojsonschema.NewStringLoader(JSONSchemaString),
		gojsonschema.NewGoLoader(&manifest),
	)
	if err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}

	if !res.Valid() {
		msg := strings.Join(kittehs.Transform(res.Errors(), func(err gojsonschema.ResultError) string {
			return fmt.Sprintf("- %s: %s", err.Field(), err.Description())
		}), "\n")
		return nil, fmt.Errorf("invalid YAML semantics:\n%s", msg)
	}

	return &manifest, nil
}
