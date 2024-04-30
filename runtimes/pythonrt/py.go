package pythonrt

import (
	"archive/tar"
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"os/exec"
	"path"
	"strings"

	"go.uber.org/zap"
)

var (
	//go:embed ak_runner.py
	runnerPyCode []byte

	//go:embed requirements.txt
	requirementsData []byte
)

func createTar(fs fs.FS) ([]byte, error) {
	var buf bytes.Buffer
	w := tar.NewWriter(&buf)
	if err := w.AddFS(fs); err != nil {
		return nil, err
	}

	w.Close()
	return buf.Bytes(), nil
}

type Version struct {
	Major int
	Minor int
}

type exeInfo struct {
	Exe           string
	VersionString string
	Version       Version
}

func parsePyVersion(s string) (major, minor int, err error) {
	// Python 3.12.2
	const prefix = "Python "
	if !strings.HasPrefix(s, prefix) {
		return 0, 0, fmt.Errorf("bad python version prefix in: %q", s)
	}

	s = s[len(prefix):]
	_, err = fmt.Sscanf(s, "%d.%d", &major, &minor)
	if err != nil {
		return 0, 0, err
	}

	return
}

func findPython() (string, error) {
	names := []string{"python3", "python"}
	for _, name := range names {
		exePath, err := exec.LookPath(name)
		if err == nil {
			return exePath, nil
		}
	}

	return "", fmt.Errorf("non of %v found in PATH", names)
}

func pyExeInfo(ctx context.Context) (exeInfo, error) {
	exePath, err := findPython()
	if err != nil {
		return exeInfo{}, err
	}

	cmd := exec.CommandContext(ctx, exePath, "--version")
	var buf bytes.Buffer
	cmd.Stdout = &buf

	if err := cmd.Run(); err != nil {
		return exeInfo{}, fmt.Errorf("%q --version: %w", exePath, err)
	}

	version := strings.TrimSpace(buf.String())
	major, minor, err := parsePyVersion(version)
	if err != nil {
		return exeInfo{}, fmt.Errorf("failed to parse Python version %q: %w", version, err)
	}

	info := exeInfo{
		Exe:           exePath,
		VersionString: version,
		Version: Version{
			Major: major,
			Minor: minor,
		},
	}
	return info, nil
}

func extractRunner(rootDir string) (string, error) {
	fileName := path.Join(rootDir, "ak_runner.py")
	file, err := os.Create(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := io.Copy(file, bytes.NewReader(runnerPyCode)); err != nil {
		return "", fmt.Errorf("can't copy python code to %s: %w", file.Name(), err)
	}

	return fileName, nil
}

type pyRunInfo struct {
	rootDir  string
	sockPath string
	lis      net.Listener
	proc     *os.Process
}

func runPython(log *zap.Logger, pyExe string, tarData []byte, rootPath string, env map[string]string, stdout, stderr io.Writer) (*pyRunInfo, error) {
	rootDir, err := os.MkdirTemp("", "ak-")
	if err != nil {
		return nil, err
	}

	tarPath, err := writeTar(rootDir, tarData)
	if err != nil {
		return nil, err
	}

	runnerPath, err := extractRunner(rootDir)
	if err != nil {
		return nil, err
	}
	log.Info("python runner", zap.String("path", runnerPath))

	sockPath := path.Join(rootDir, "ak.sock")
	log.Info("socket", zap.String("path", sockPath))

	lis, err := net.Listen("unix", sockPath)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(pyExe, runnerPath, sockPath, tarPath, rootPath)
	cmd.Dir = rootDir
	cmd.Env = overrideEnv(env)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Start(); err != nil {
		lis.Close()
		return nil, err
	}

	info := pyRunInfo{
		rootDir:  rootPath,
		sockPath: sockPath,
		lis:      lis,
		proc:     cmd.Process,
	}

	return &info, nil
}

func writeTar(rootDir string, data []byte) (string, error) {
	tarName := path.Join(rootDir, "code.tar")
	file, err := os.Create(tarName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := io.Copy(file, bytes.NewReader(data)); err != nil {
		return "", err
	}

	return tarName, err
}

func overrideEnv(envMap map[string]string) []string {
	env := os.Environ()
	// Append AK values to end to override (see Env docs in https://pkg.go.dev/os/exec#Cmd)
	for k, v := range envMap {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	return env
}

func createVEnv(pyExe string, venvPath string) error {
	cmd := exec.Command(pyExe, "-m", "venv", venvPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("create venv: %w", err)
	}

	file, err := os.CreateTemp("", "")
	if err != nil {
		return err
	}

	if _, err := io.Copy(file, bytes.NewReader(requirementsData)); err != nil {
		file.Close()
		return fmt.Errorf("copy requirements to %q: %w", file.Name(), err)
	}
	file.Close()

	venvPy := path.Join(venvPath, "bin", "python")
	cmd = exec.Command(venvPy, "-m", "pip", "install", "-r", file.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("install dependencies from %q: %w", file.Name(), err)
	}

	return nil
}
