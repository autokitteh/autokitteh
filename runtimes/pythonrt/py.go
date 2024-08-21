package pythonrt

import (
	"archive/tar"
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/google/uuid"
	"go.autokitteh.dev/autokitteh/runtimes/pythonrt/pb"

	"go.uber.org/zap"
)

var (
	//go:embed runner/*.py
	runnerPyCode embed.FS

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

// findPython finds `python3` or `python` in PATH.
func findPython() (string, error) {
	names := []string{"python3", "python"}
	for _, name := range names {
		exePath, err := exec.LookPath(name)
		if err == nil {
			return exePath, nil
		}
	}

	return "", fmt.Errorf("none of %v found in PATH", names)
}

func pyExeInfo(ctx context.Context, exePath string) (exeInfo, error) {
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

type pyRunInfo struct {
	userRootDir string
	pyRootDir   string
	client      pb.RunnerClient
	proc        *os.Process
	port        int
}

func (ri *pyRunInfo) Cleanup() {
	if ri.userRootDir != "" {
		os.RemoveAll(ri.userRootDir)
	}

	if ri.pyRootDir != "" {
		os.RemoveAll(ri.pyRootDir)
	}
}

type runOptions struct {
	log            *zap.Logger
	pyExe          string // Python executable to use
	entryPoint     string // simple.py:greet
	tarData        []byte
	env            map[string]string // Python process environment
	stdout, stderr io.Writer
	workerAddr     string
}

func runPython(opts runOptions) (*pyRunInfo, error) {
	runOK := false
	var ri pyRunInfo
	var err error
	defer func() {
		if !runOK {
			ri.Cleanup()
		}
	}()

	ri.port, err = freePort()
	if err != nil {
		return nil, fmt.Errorf("cannot find free port: %w", err)
	}

	ri.userRootDir, err = os.MkdirTemp("", "ak-")
	if err != nil {
		return nil, err
	}
	opts.log.Info("user root dir", zap.String("path", ri.userRootDir))

	if err := extractTar(ri.userRootDir, opts.tarData); err != nil {
		return nil, err
	}

	ri.pyRootDir, err = os.MkdirTemp("", "ak-")
	if err != nil {
		return nil, err
	}
	opts.log.Info("python root dir", zap.String("path", ri.pyRootDir))

	if err := copyFS(runnerPyCode, ri.pyRootDir); err != nil {
		return nil, err
	}

	mainPy := path.Join(ri.pyRootDir, "main.py")
	cmd := exec.Command(
		opts.pyExe, "-u", mainPy,
		"--worker-address", opts.workerAddr,
		"--port", fmt.Sprintf("%d", ri.port),
		"--runner-id", uuid.NewString(),
		"--code-dir", ri.pyRootDir,
	)
	// cmd.Dir = ri.pyRootDir # TODO: Check if required
	cmd.Env = overrideEnv(opts.env, ri.pyRootDir, ri.userRootDir)
	cmd.Stdout = opts.stdout
	cmd.Stderr = opts.stderr

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	runOK = true // signal we're good to cleanup
	ri.proc = cmd.Process

	return &ri, nil
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

type Export struct {
	Name string
	File string
	Line int
}

func pyExports(ctx context.Context, pyExe string, fsys fs.FS) ([]Export, error) {
	tmpDir, err := os.MkdirTemp("", "ak-inspect-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	if err := copyFS(fsys, tmpDir); err != nil {
		return nil, err
	}

	runnerDir, err := os.MkdirTemp("", "ak-inspect-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(runnerDir)

	if err := copyFS(runnerPyCode, runnerDir); err != nil {
		return nil, err
	}

	outFile, err := os.CreateTemp("", "")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	outFile.Close()

	cmd := exec.CommandContext(ctx, pyExe, "-m", runnerMod, "inspect", tmpDir, outFile.Name())
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	cmd.Env = adjustPythonPath(os.Environ(), runnerDir)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("inspect: %w.\nPython output: %s", err, buf.String())
	}

	file, err := os.Open(outFile.Name())
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var exports []Export
	if err := json.NewDecoder(file).Decode(&exports); err != nil {
		return nil, err
	}

	return exports, nil
}
