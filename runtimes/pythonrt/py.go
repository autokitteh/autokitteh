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
	"strings"

	"go.uber.org/zap"
)

var (
	//go:embed ak_runner.py
	runnerPyCode []byte
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

type execInfo struct {
	Exe     string
	Version string
}

func pyExecInfo(ctx context.Context) (execInfo, error) {
	exePath, err := exec.LookPath("python")
	if err != nil {
		return execInfo{}, err
	}

	cmd := exec.CommandContext(ctx, exePath, "--version")
	var buf bytes.Buffer
	cmd.Stdout = &buf

	if err := cmd.Run(); err != nil {
		return execInfo{}, fmt.Errorf("%q --version: %w", exePath, err)
	}

	version := strings.TrimSpace(buf.String())
	return execInfo{Exe: exePath, Version: version}, nil
}

func extractRunner() (string, error) {
	file, err := os.CreateTemp("", "*.py")
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := io.Copy(file, bytes.NewReader(runnerPyCode)); err != nil {
		return "", fmt.Errorf("can't copy python code to %s - %w", file.Name(), err)
	}

	return file.Name(), nil
}

type pyRunInfo struct {
	sockPath string
	lis      net.Listener
	proc     *os.Process
}

func newSocket() (string, error) {
	file, err := os.CreateTemp("", "*.sock")
	if err != nil {
		return "", err
	}
	file.Close()

	os.RemoveAll(file.Name())
	return file.Name(), nil
}

func runPython(log *zap.Logger, tarData []byte, rootPath string, env map[string]string) (*pyRunInfo, error) {
	tarPath, err := writeTar(tarData)
	if err != nil {
		return nil, err
	}

	runnerPath, err := extractRunner()
	if err != nil {
		return nil, err
	}
	log.Info("python runner", zap.String("path", runnerPath))

	sockPath, err := newSocket()
	if err != nil {
		return nil, err
	}
	log.Info("socket", zap.String("path", sockPath))

	lis, err := net.Listen("unix", sockPath)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("python", runnerPath, sockPath, tarPath, rootPath)
	// TODO: Hook cmd.Stdout & cmd.Stderr to logs (ENG-552)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = overrideEnv(env)

	if err := cmd.Start(); err != nil {
		lis.Close()
		return nil, err
	}

	info := pyRunInfo{
		sockPath: sockPath,
		lis:      lis,
		proc:     cmd.Process,
	}

	return &info, nil
}

func writeTar(data []byte) (string, error) {
	file, err := os.CreateTemp("", "*.tar")
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := io.Copy(file, bytes.NewReader(data)); err != nil {
		return "", err
	}

	return file.Name(), err
}

func overrideEnv(envMap map[string]string) []string {
	env := os.Environ()
	// Append AK values to end to override (see Env docs in https://pkg.go.dev/os/exec#Cmd)
	for k, v := range envMap {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	return env
}
