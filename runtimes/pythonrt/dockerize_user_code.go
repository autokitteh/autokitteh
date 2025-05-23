package pythonrt

import (
	_ "embed"
	"io/fs"
	"os"
	"path"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/backend/tar"
)

//go:embed dockerfilenodeps
var dockerfileNoDeps string

//go:embed dockerfilewithdeps
var dockerfileWithDeps string

//go:embed sitecustomize.py
var siteCustomize []byte

func prepareUserCode(code []byte, gzipped bool) (string, error) {
	tf, err := tar.FromBytes(code, gzipped)
	if err != nil {
		return "", err
	}

	content, err := tf.Content()
	if err != nil {
		return "", err
	}

	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", err
	}

	pycode, err := fs.Sub(runnerPyCode, "runner")
	if err != nil {
		return "", err
	}
	runnerDir := path.Join(tmpDir, "/runner")
	if err := os.Mkdir(runnerDir, 0o777); err != nil {
		return "", err
	}

	if _, err := copyFS(pycode, runnerDir); err != nil {
		return "", err
	}

	workflowDir := path.Join(tmpDir, "/workflow")

	if err := os.Mkdir(workflowDir, 0o777); err != nil {
		return "", err
	}

	hasRequirementsFile := false
	for file, content := range content {
		if strings.HasPrefix(file, ".") {
			continue
		}

		dir := path.Dir(path.Join(workflowDir, file))
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return "", err
		}

		if file == "requirements.txt" {
			hasRequirementsFile = true
			if err := os.WriteFile(path.Join(workflowDir, "user_requirements.txt"), content, 0o777); err != nil {
				return "", err
			}
		} else {
			if err := os.WriteFile(path.Join(workflowDir, file), content, 0o777); err != nil {
				return "", err
			}
		}

	}

	if err := os.WriteFile(path.Join(tmpDir, "sitecustomize.py"), siteCustomize, 0o777); err != nil {
		return "", err
	}

	dockerfile := []byte(dockerfileNoDeps)
	if hasRequirementsFile {
		dockerfile = []byte(dockerfileWithDeps)
	}
	if err := os.WriteFile(path.Join(tmpDir, "Dockerfile"), dockerfile, 0o777); err != nil {
		return "", err
	}

	if err := os.WriteFile(path.Join(tmpDir, "pyproject.toml"), pyProjectTOML, 0o777); err != nil {
		return "", err
	}

	if _, err := copyFS(pysdk, tmpDir); err != nil {
		return "", err
	}

	return tmpDir, nil
}
