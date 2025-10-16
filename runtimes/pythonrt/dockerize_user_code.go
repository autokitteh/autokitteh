package pythonrt

import (
	_ "embed"
	"io/fs"
	"os"
	"path"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/backend/tar"
)

//go:embed Dockerfile
var dockerfile string

//go:embed sitecustomize.py
var siteCustomize []byte

func prepareBaseImageCode(customReqFilePath string) (string, error) {
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
	customReqBytes := []byte("")
	if customReqFilePath != "" {
		customReqBytes, err = os.ReadFile(customReqFilePath)
		if err != nil {
			return "", err
		}
	}

	if err := os.WriteFile(path.Join(workflowDir, "user_requirements.txt"), customReqBytes, 0o777); err != nil {
		return "", err
	}

	if err := os.WriteFile(path.Join(tmpDir, "sitecustomize.py"), siteCustomize, 0o777); err != nil {
		return "", err
	}

	if err := os.WriteFile(path.Join(tmpDir, "Dockerfile"), []byte(dockerfile), 0o777); err != nil {
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

type userCodeDetails struct {
	codePath              string
	hasCustomRequirements bool
	requirementsFilePath  string
}

func writeUserCodeToFS(files map[string][]byte) (userCodeDetails, error) {
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return userCodeDetails{}, err
	}

	if err := os.Chmod(tmpDir, 0o777); err != nil {
		return userCodeDetails{}, err
	}

	hasRequirementsFile := false
	for file, content := range files {
		if strings.HasPrefix(file, ".") || strings.HasSuffix(file, "/") {
			continue
		}

		dir := path.Dir(path.Join(tmpDir, file))
		if err := os.MkdirAll(dir, 0o777); err != nil {
			return userCodeDetails{}, err
		}

		if file == "requirements.txt" {
			hasRequirementsFile = true
			if err := os.WriteFile(path.Join(tmpDir, "user_requirements.txt"), content, 0o644); err != nil {
				return userCodeDetails{}, err
			}
		} else {
			targetPath := path.Join(tmpDir, file)
			if err := os.WriteFile(targetPath, content, 0o777); err != nil {
				return userCodeDetails{}, err
			}
		}
	}

	return userCodeDetails{
		codePath:              tmpDir,
		hasCustomRequirements: hasRequirementsFile,
		requirementsFilePath:  path.Join(tmpDir, "user_requirements.txt"),
	}, nil
}

func prepareUserCode(code []byte, gzipped bool) (userCodeDetails, error) {
	tf, err := tar.FromBytes(code, gzipped)
	if err != nil {
		return userCodeDetails{}, err
	}

	files, err := tf.Content()
	if err != nil {
		return userCodeDetails{}, err
	}

	return writeUserCodeToFS(files)
}
