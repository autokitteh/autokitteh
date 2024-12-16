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
	"path/filepath"
	"strings"
	"syscall"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.jetify.com/typeid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapio"
)

var (
	//go:embed all:runner
	runnerPyCode embed.FS

	//go:embed all:py-sdk
	pysdk embed.FS

	//go:embed runner/pyproject.toml
	pyProjectTOML []byte
)

type LocalPython struct {
	log                *zap.Logger
	userDir            string
	runnerDir          string
	port               int
	proc               *os.Process
	id                 string
	logRunnerCode      bool
	sessionID          sdktypes.SessionID
	stdoutRunnerLogger *zapio.Writer
	stderrRunnerLogger *zapio.Writer
}

func (r *LocalPython) Close() error {
	var err error

	if r.proc != nil {
		if kerr := r.proc.Kill(); kerr != nil {
			err = errors.Join(err, fmt.Errorf("kill runner (pid=%d) - %w", r.proc.Pid, kerr))
		}

		_, waitErr := r.proc.Wait()
		if waitErr != nil {
			err = errors.Join(err, fmt.Errorf("wait runner (pid=%d) - %w", r.proc.Pid, waitErr))
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

	if r.stdoutRunnerLogger != nil {
		if rlerr := r.stdoutRunnerLogger.Close(); rlerr != nil {
			err = errors.Join(err, fmt.Errorf("close stdout runner logger - %w", rlerr))
		}
	}

	if r.stderrRunnerLogger != nil {
		if rlerr := r.stderrRunnerLogger.Close(); rlerr != nil {
			err = errors.Join(err, fmt.Errorf("close stderr runner logger - %w", rlerr))
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

	id, err := typeid.WithPrefix("runner")
	if err != nil {
		return err
	}
	r.id = id.String()

	mainPy := path.Join(r.runnerDir, "runner", "main.py")
	cmd := exec.Command(
		pyExe, "-u", mainPy,
		"--worker-address", workerAddr,
		"--port", fmt.Sprintf("%d", r.port),
		"--runner-id", r.id,
		"--code-dir", r.userDir,
	)
	cmd.Env = overrideEnv(env, r.runnerDir)
	cmd.Dir = r.userDir

	if r.logRunnerCode {
		r.stdoutRunnerLogger = &zapio.Writer{Log: r.log.With(zap.String("stream", "stdout")), Level: zap.InfoLevel}
		cmd.Stdout = r.stdoutRunnerLogger
		// Why warn instead of error? (1) We're using stdout too much, (2) Python errors are not
		// necessarily AK errors, and (3) errors include a stack trace, which is irrelevant here.
		r.stderrRunnerLogger = &zapio.Writer{Log: r.log.With(zap.String("stream", "stderr")), Level: zap.WarnLevel}
		cmd.Stderr = r.stderrRunnerLogger
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
	var status syscall.WaitStatus

	pid, err := syscall.Wait4(r.proc.Pid, &status, syscall.WNOHANG, nil)
	if err != nil {
		return fmt.Errorf("wait proc: %w", err)
	}

	if pid == r.proc.Pid { // state changed
		if status.Signaled() {
			sig := status.Signal()
			switch sig {
			case syscall.SIGKILL, syscall.SIGTERM, syscall.SIGABRT, syscall.SIGSEGV, syscall.SIGBUS:
				return fmt.Errorf("proc signaled (%v)", sig)
			default:
				// maybe process communicates with itself? do nothing meanwhile
				return nil
			}
		}
		if status.Exited() {
			return fmt.Errorf("proc exited with %d", status.ExitStatus())
		}
	}

	err = r.proc.Signal(syscall.Signal(0))
	if err != nil {
		if _, err := os.FindProcess(r.proc.Pid); err != nil {
			return fmt.Errorf("no runner proc found. %w ", err)
		}
	}
	return err
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

// extractTar extracts a project deployment tar file into a random directory.
func extractTar(rootDir string, data []byte) error {
	rootDir = filepath.Clean(rootDir)
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

		outPath := filepath.Clean(path.Join(rootDir, hdr.Name))

		// Prevent "zip slips": we generate the tar file, but better safe than sorry.
		if !strings.HasPrefix(outPath, rootDir) {
			return fmt.Errorf("extracted file %q is outside root dir", hdr.Name)
		}

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

func overrideEnv(envMap map[string]string, runnerPath string) []string {
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

	venvPy := path.Join(venvPath, "bin", "python")
	if err := install(venvPy, tmpDir, []string{"-m", "pip", "install", ".[all]"}); err != nil {
		return fmt.Errorf("install dependencies from %q: %w", file.Name(), err)
	}

	if err := copyFS(pysdk, tmpDir); err != nil {
		return err
	}

	if err = install(venvPy, path.Join(tmpDir, "/py-sdk"), []string{"-m", "pip", "install", "."}); err != nil {
		return fmt.Errorf("install autokitteh py sdk %w", err)
	}
	return nil
}

func install(pyPath, cwd string, args []string) error {
	cmd := exec.Command(pyPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = cwd

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
