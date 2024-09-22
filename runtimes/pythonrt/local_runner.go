package pythonrt

import (
	"archive/tar"
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	//go:embed runner/*.py
	runnerPyCode embed.FS

	//go:embed runner/pyproject.toml
	pyProjectTOML []byte
)

type LocalPython struct {
	log           *zap.Logger
	userDir       string
	runnerDir     string
	port          int
	proc          *os.Process
	id            string
	logRunnerCode bool
}

func (r *LocalPython) Close() error {
	var err error

	if r.proc != nil {
		if kerr := r.proc.Kill(); kerr != nil {
			err = errors.Join(err, fmt.Errorf("kill runner (pid=%d) - %w", r.proc.Pid, kerr))
		}
	}

	if r.userDir != "" {
		if uerr := os.RemoveAll(r.userDir); uerr != nil {
			err = errors.Join(err, fmt.Errorf("clean user dir %q - %w", r.userDir, uerr))
		}
	}

	if r.runnerDir != "" {
		if perr := os.RemoveAll(r.runnerDir); perr != nil {
			err = errors.Join(err, fmt.Errorf("clean runner dir %q - %w", r.runnerDir, perr))
		}
	}

	return err
}

func freePort() (int, error) {
	conn, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}

	conn.Close()
	return conn.Addr().(*net.TCPAddr).Port, nil
}

func (r *LocalPython) Start(pyExe string, tarData []byte, env map[string]string, workerAddr string) error {
	runOK := false

	defer func() {
		if !runOK {
			if err := r.Close(); err != nil {
				r.log.Warn("cleanup runner", zap.Error(err))
			}
		}
	}()

	port, err := freePort()
	if err != nil {
		return fmt.Errorf("cannot find free port: %w", err)
	}
	r.port = port

	userDir, err := os.MkdirTemp("", "ak-user-")
	if err != nil {
		return fmt.Errorf("create user directory - %w", err)
	}
	r.userDir = userDir
	r.log.Info("user root dir", zap.String("path", r.userDir))

	if err := extractTar(r.userDir, tarData); err != nil {
		return fmt.Errorf("extract user tar - %w", err)
	}

	runnerDir, err := os.MkdirTemp("", "ak-runner-")
	if err != nil {
		return fmt.Errorf("create runner dir - %w", err)
	}
	r.runnerDir = runnerDir
	r.log.Info("python root dir", zap.String("path", r.runnerDir))

	if err := copyFS(runnerPyCode, r.runnerDir); err != nil {
		return fmt.Errorf("copy runner code - %w", err)
	}

	r.id = uuid.NewString()
	mainPy := path.Join(r.runnerDir, "runner", "main.py")
	cmd := exec.Command(
		pyExe, "-u", mainPy,
		"--worker-address", workerAddr,
		"--port", fmt.Sprintf("%d", r.port),
		"--runner-id", r.id,
		"--code-dir", r.userDir,
	)
	cmd.Env = overrideEnv(env, r.runnerDir, r.userDir)
	if r.logRunnerCode {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	// make sure runner is killed if ak is killed
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start runner - %w", err)
	}

	runOK = true // signal we're good to cleanup
	r.proc = cmd.Process
	r.log.Info("started python runner", zap.String("command", cmd.String()), zap.Int("pid", r.proc.Pid))
	return nil
}

func (r *LocalPython) Health() error {
	return r.proc.Signal(syscall.Signal(0))
}

func createTar(fs fs.FS) ([]byte, error) {
	var buf bytes.Buffer
	w := tar.NewWriter(&buf)
	if err := w.AddFS(fs); err != nil {
		return nil, err
	}

	w.Close()
	return buf.Bytes(), nil
}

// Copy fs to file so Python can inspect
// TODO: Once os.CopyFS makes it out we can remove this
// https://github.com/golang/go/issues/62484
func copyFS(fsys fs.FS, root string) error {
	return fs.WalkDir(fsys, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !entry.Type().IsRegular() {
			return nil
		}

		dest := path.Join(root, name)
		dirName := path.Dir(dest)
		if err := os.MkdirAll(dirName, 0o755); err != nil {
			return err
		}

		r, err := fsys.Open(name)
		if err != nil {
			return err
		}
		defer r.Close()

		w, err := os.Create(dest)
		if err != nil {
			return err
		}
		defer w.Close()

		if _, err := io.Copy(w, r); err != nil {
			return err
		}

		return nil
	})
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

func extractTar(rootDir string, data []byte) error {
	rdr := tar.NewReader(bytes.NewReader(data))
	for {
		hdr, err := rdr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if hdr.FileInfo().IsDir() {
			continue
		}

		outPath := path.Join(rootDir, hdr.Name)
		if err := os.MkdirAll(path.Dir(outPath), 0o755); err != nil {
			return err
		}

		file, err := os.Create(outPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(file, rdr)
		if err != nil {
			return err
		}
	}

	return nil
}

func adjustPythonPath(env []string, runnerPath string) []string {
	// Iterate in reverse since last value overrides
	for i := len(env) - 1; i >= 0; i-- {
		v := env[i]
		if strings.HasPrefix(v, "PYTHONPATH=") {
			env[i] = fmt.Sprintf("%s:%s", v, runnerPath)
			return env
		}
	}

	return append(env, fmt.Sprintf("PYTHONPATH=%s", runnerPath))
}

func overrideEnv(envMap map[string]string, runnerPath, userCodePath string) []string {
	env := os.Environ()
	// Append AK values to end to override (see Env docs in https://pkg.go.dev/os/exec#Cmd)
	for k, v := range envMap {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	return adjustPythonPath(env, runnerPath)
}

func createVEnv(pyExe string, venvPath string) error {
	cmd := exec.Command(pyExe, "-m", "venv", venvPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("create venv: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "ak-proj")
	if err != nil {
		return err
	}

	outFile := path.Join(tmpDir, "pyproject.toml")
	file, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.Copy(file, bytes.NewReader(pyProjectTOML)); err != nil {
		return fmt.Errorf("copy requirements to %q: %w", file.Name(), err)
	}
	file.Close()

	venvPy := path.Join(venvPath, "bin", "python")
	cmd = exec.Command(venvPy, "-m", "pip", "install", ".[all]")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = tmpDir

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("install dependencies from %q: %w", file.Name(), err)
	}

	return nil
}
