package nodejsrt

import (
	"archive/tar"
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"go.jetify.com/typeid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapio"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	//go:embed all:nodejs/runtime
	runnerJsCode embed.FS
)

type LocalNodeJS struct {
	log                *zap.Logger
	projectDir         string
	port               int
	cmd                *exec.Cmd // Store the full command instead of just the process
	id                 string
	logRunnerCode      bool
	sessionID          sdktypes.SessionID
	stdoutRunnerLogger *zapio.Writer
	stderrRunnerLogger *zapio.Writer
	cleanupWorkspace   bool // If true, workspace will be removed on Close
}

// NewLocalNodeJS creates a new LocalNodeJS instance with the given logger
func NewLocalNodeJS(logger *zap.Logger) *LocalNodeJS {
	return &LocalNodeJS{
		log: logger,
	}
}

func (r *LocalNodeJS) Close() error {
	var err error

	if r.cmd != nil && r.cmd.Process != nil {
		// First try SIGTERM for graceful shutdown
		if terr := r.cmd.Process.Signal(syscall.SIGTERM); terr != nil {
			r.log.Warn("SIGTERM failed", zap.Error(terr))
		}

		// Give it a moment to shut down gracefully
		done := make(chan error)
		go func() {
			done <- r.cmd.Wait()
		}()

		// Wait for process to exit or timeout
		select {
		case <-time.After(3 * time.Second):
			// Force kill if still running
			if kerr := r.cmd.Process.Kill(); kerr != nil {
				err = errors.Join(err, fmt.Errorf("kill process (pid=%d) - %w", r.cmd.Process.Pid, kerr))
			}
		case werr := <-done:
			if werr != nil && !strings.Contains(werr.Error(), "signal: terminated") {
				err = errors.Join(err, fmt.Errorf("wait for process (pid=%d) - %w", r.cmd.Process.Pid, werr))
			}
		}
	}

	if r.projectDir != "" && r.cleanupWorkspace {
		if uerr := os.RemoveAll(r.projectDir); uerr != nil {
			err = errors.Join(err, fmt.Errorf("clean project dir %q - %w", r.projectDir, uerr))
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

// PrepareWorkspace prepares the runtime workspace. Creates a new temporary workspace if no directory was set.
func (r *LocalNodeJS) PrepareWorkspace() error {
	if strings.TrimSpace(r.projectDir) != "" {
		r.log.Info("using provided project dir", zap.String("path", r.projectDir))
		return nil
	}

	// Create temporary directory for the project
	tempDir, err := os.MkdirTemp("", "ak-project-")
	if err != nil {
		return fmt.Errorf("create project directory - %w", err)
	}
	r.projectDir = tempDir
	r.cleanupWorkspace = true
	r.log.Info("created temporary project dir", zap.String("path", r.projectDir))
	return nil
}

// PrepareProject extracts the tar and sets up the runner code in r.projectDir
func (r *LocalNodeJS) PrepareProject(tarData []byte) error {
	if strings.TrimSpace(r.projectDir) == "" {
		return fmt.Errorf("project directory not set - call PrepareWorkspace first")
	}

	if err := extractTar(r.projectDir, tarData); err != nil {
		return fmt.Errorf("extract user tar - %w", err)
	}
	r.log.Info("extracted tar to project dir")

	akDir := path.Join(r.projectDir, ".ak")
	if err := os.MkdirAll(akDir, 0o755); err != nil {
		return fmt.Errorf("create .ak directory - %w", err)
	}

	nodejsFS, err := fs.Sub(runnerJsCode, "nodejs")
	if err != nil {
		return fmt.Errorf("get nodejs subdir: %w", err)
	}
	if err := copyFSToDir(nodejsFS, akDir); err != nil {
		return fmt.Errorf("copy runner code - %w", err)
	}
	r.log.Info("copied runner code to .ak directory")

	return nil
}

func (r *LocalNodeJS) setupDependencies() error {
	//if err := setupTypeScript(r.projectDir); err != nil {
	//	return fmt.Errorf("setup typescript: %w", err)
	//}
	//r.log.Info("typescript setup complete")

	if err := installDependencies(r.projectDir); err != nil {
		return fmt.Errorf("install dependencies: %w", err)
	}
	r.log.Info("dependencies installed")

	return nil
}

// SetProjectDir sets the project directory for the runner.
// If the directory doesn't exist, it will be created.
// When a directory is set, it won't be cleaned up on Close.
func (r *LocalNodeJS) SetProjectDir(dir string) error {
	if strings.TrimSpace(dir) == "" {
		return fmt.Errorf("project directory cannot be empty")
	}

	dir = filepath.Clean(dir)

	// Check if directory exists
	info, err := os.Stat(dir)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("check directory %q: %w", dir, err)
		}
		// Directory doesn't exist, create it
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create directory %q: %w", dir, err)
		}
		r.log.Info("created project directory", zap.String("path", dir))
	} else {
		// Directory exists, verify it's a directory
		if !info.IsDir() {
			return fmt.Errorf("path %q exists but is not a directory", dir)
		}
		// Check if we have write permission by attempting to create a temporary file
		testFile := filepath.Join(dir, ".ak.test")
		f, err := os.Create(testFile)
		if err != nil {
			return fmt.Errorf("directory %q is not writable: %w", dir, err)
		}
		f.Close()
		os.Remove(testFile)
	}

	r.projectDir = dir
	r.cleanupWorkspace = false
	r.log.Info("using project directory", zap.String("path", dir))
	return nil
}

// Start prepares the workspace, project, dependencies and runs the node process.
func (r *LocalNodeJS) Start(nodeExe, tarData []byte, env map[string]string, workerAddr string) error {
	// Prepare the workspace (either temporary or provided)
	if err := r.PrepareWorkspace(); err != nil {
		return fmt.Errorf("prepare workspace: %w", err)
	}

	// Prepare the project in the workspace
	if err := r.PrepareProject(tarData); err != nil {
		return fmt.Errorf("prepare project: %w", err)
	}

	// Setup dependencies
	if err := r.setupDependencies(); err != nil {
		return fmt.Errorf("setup dependencies: %w", err)
	}

	// Finally run the node process
	if err := r.RunNode(workerAddr); err != nil {
		return fmt.Errorf("run node: %w", err)
	}

	return nil
}

// RunNode starts the Node.js process in the prepared workspace.
// This should only be called after PrepareWorkspace.
func (r *LocalNodeJS) RunNode(workerAddr string) error {
	//absPath, err := filepath.Abs("nodejs/testdata/simple-test/runtime")
	//r.projectDir = absPath

	if strings.TrimSpace(r.projectDir) == "" {
		return fmt.Errorf("project directory not set - call PrepareWorkspace first")
	}

	port, err := freePort()
	if err != nil {
		return fmt.Errorf("cannot find free port: %w", err)
	}
	r.port = port
	r.log.Info("found free port", zap.Int("port", port))
	r.log.Info("projectDir", zap.String("path", r.projectDir))

	id, err := typeid.WithPrefix("runner")
	if err != nil {
		return err
	}
	r.id = id.String()

	cmd := exec.Command(
		"npx",
		"ts-node",
		".ak/runtime/runner/main.ts",
		"--worker-address", workerAddr,
		"--port", strconv.Itoa(r.port),
		"--runner-id", r.id,
		"--code-dir", r.projectDir,
	)
	cmd.Dir = r.projectDir

	if r.logRunnerCode {
		r.stdoutRunnerLogger = &zapio.Writer{Log: r.log.With(zap.String("stream", "stdout")), Level: zap.InfoLevel}
		cmd.Stdout = r.stdoutRunnerLogger
		r.stderrRunnerLogger = &zapio.Writer{Log: r.log.With(zap.String("stream", "stderr")), Level: zap.WarnLevel}
		cmd.Stderr = r.stderrRunnerLogger
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start runner - %w", err)
	}

	r.cmd = cmd
	r.log.Info("started nodejs runner",
		zap.String("command", cmd.String()),
		zap.Int("pid", cmd.Process.Pid))
	return nil
}

func (r *LocalNodeJS) Health() error {
	var status syscall.WaitStatus

	pid, err := syscall.Wait4(r.cmd.Process.Pid, &status, syscall.WNOHANG, nil)
	if err != nil {
		return fmt.Errorf("wait proc: %w", err)
	}

	if pid == r.cmd.Process.Pid {
		if status.Signaled() {
			sig := status.Signal()
			switch sig {
			case syscall.SIGKILL, syscall.SIGTERM, syscall.SIGABRT, syscall.SIGSEGV, syscall.SIGBUS:
				return fmt.Errorf("proc signaled (%v)", sig)
			default:
				return nil
			}
		}
		if status.Exited() {
			return fmt.Errorf("proc exited with %d", status.ExitStatus())
		}
	}

	err = r.cmd.Process.Signal(syscall.Signal(0))
	if err != nil {
		if _, err := os.FindProcess(r.cmd.Process.Pid); err != nil {
			return fmt.Errorf("no runner proc found. %w ", err)
		}
	}
	return err
}

func createTar(fsys fs.FS) ([]byte, error) {
	var buf bytes.Buffer
	w := tar.NewWriter(&buf)
	defer w.Close()

	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Get file info
		info, err := d.Info()
		if err != nil {
			return err
		}

		// Skip symbolic links
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = path

		// Write header
		if err := w.WriteHeader(header); err != nil {
			return err
		}

		// If it's a directory, we're done
		if info.IsDir() {
			return nil
		}

		// Copy file content
		file, err := fsys.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(w, file)
		return err
	})

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func copyFSToDir(fsys fs.FS, root string) error {
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

// setupTypeScript ensures proper TypeScript configuration
func setupTypeScript(projectDir string) error {
	// Create .ak/types directory
	typesDir := filepath.Join(projectDir, ".ak", "types")
	if err := os.MkdirAll(typesDir, 0o755); err != nil {
		return fmt.Errorf("create types directory: %w", err)
	}

	// Create global.d.ts
	globalDTS := `declare global {
    /**
     * Autokitteh activity call function
     * @param activityName Name of the activity to call
     * @param args Arguments for the activity
     * @returns Promise with the activity result
     */
    function ak_call(activityName: string, ...args: unknown[]): Promise<unknown>;
}

export {};`

	if err := os.WriteFile(filepath.Join(typesDir, "global.d.ts"), []byte(globalDTS), 0644); err != nil {
		return fmt.Errorf("write global.d.ts: %w", err)
	}

	// Create package.json if it doesn't exist
	pkgPath := filepath.Join(projectDir, "package.json")
	if _, err := os.Stat(pkgPath); os.IsNotExist(err) {
		minimalPkg := `{
  "name": "ak-project",
  "version": "1.0.0",
  "private": true
}`
		if err := os.WriteFile(pkgPath, []byte(minimalPkg), 0644); err != nil {
			return fmt.Errorf("write package.json: %w", err)
		}
	}

	// Setup tsconfig.json
	tsConfig := map[string]interface{}{
		"compilerOptions": map[string]interface{}{
			"typeRoots": []string{
				"node_modules/@types",
				".ak/types",
			},
			"esModuleInterop":  true,
			"skipLibCheck":     true,
			"target":           "ES2020",
			"module":           "CommonJS",
			"moduleResolution": "node",
			"sourceMap":        true,
			"outDir":           "dist",
			"strict":           true,
		},
	}

	// Read existing tsconfig if it exists
	tsconfigPath := filepath.Join(projectDir, "tsconfig.json")
	if existingConfig, err := os.ReadFile(tsconfigPath); err == nil {
		var existing map[string]interface{}
		if err := json.Unmarshal(existingConfig, &existing); err == nil {
			// Merge with existing config
			tsConfig = mergeConfigs(existing, tsConfig)
		}
	}

	// Write back tsconfig
	tsconfigBytes, _ := json.MarshalIndent(tsConfig, "", "  ")
	if err := os.WriteFile(tsconfigPath, tsconfigBytes, 0644); err != nil {
		return fmt.Errorf("write tsconfig.json: %w", err)
	}

	return nil
}

// mergeConfigs merges two configuration objects
func mergeConfigs(existing, new map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range existing {
		result[k] = v
	}
	for k, v := range new {
		if existingVal, ok := result[k]; ok {
			if existingMap, ok := existingVal.(map[string]interface{}); ok {
				if newMap, ok := v.(map[string]interface{}); ok {
					result[k] = mergeConfigs(existingMap, newMap)
					continue
				}
			}
		}
		result[k] = v
	}
	return result
}

// installDependencies installs both user and ak dependencies
func installDependencies(projectDir string) error {
	// Helper function for npm commands
	runNpm := func(args ...string) error {
		cmd := exec.Command("npm", args...)
		cmd.Dir = projectDir

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("npm command failed: %w\nstdout: %s\nstderr: %s",
				err, stdout.String(), stderr.String())
		}
		return nil
	}

	// First install user's dependencies
	if err := runNpm("install"); err != nil {
		return fmt.Errorf("user dependencies install failed: %w", err)
	}

	// Install ak framework dependencies
	pkgPath := filepath.Join(projectDir, ".ak/runtime", "package.json")
	pkgData, err := os.ReadFile(pkgPath)
	if err != nil {
		return fmt.Errorf("read .ak/package.json: %w", err)
	}

	var pkgJSON struct {
		Dependencies map[string]string `json:"dependencies"`
	}

	if err := json.Unmarshal(pkgData, &pkgJSON); err != nil {
		return fmt.Errorf("parse .ak/package.json: %w", err)
	}

	// Install ak dependencies from .ak/package.json
	for pkg, version := range pkgJSON.Dependencies {
		if err := runNpm("install", "--save", pkg+"@"+version); err != nil {
			return fmt.Errorf("failed to install %s: %w", pkg, err)
		}
	}

	return nil
}

// SetupDependencies installs TypeScript and project dependencies
func (r *LocalNodeJS) SetupDependencies() error {
	return r.setupDependencies()
}
