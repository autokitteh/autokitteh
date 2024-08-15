package pythonrt

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	pb "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/remote/v1"
)

type localRunnerManager struct {
	logger           *zap.Logger
	pyExe            string
	runnerIDToRunner map[string]*LocalPython
	mu               *sync.Mutex
	workerAddress    string
	cfg              LocalRunnerConfig
}

type LocalRunnerConfig struct {
	WorkerAddress string
	LazyLoadVEnv  bool
}

func ConfigureLocalRunnerManager(log *zap.Logger, cfg LocalRunnerConfig) error {
	// if configuredRunnerType != runnerTypeNotConfigured {
	// 	if configuredRunnerType == runnerTypeLocal {
	// 		return nil
	// 	}
	// 	return errors.New("runner type already configured, cannot configure twice")
	// }

	pyExe, isUserPy, err := pythonToRun(log)
	if err != nil {
		return err
	}

	const timeout = 3 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	info, err := pyExeInfo(ctx, pyExe)
	if err != nil {
		return err
	}

	log.Info("python info", zap.String("exe", info.Exe), zap.Any("version", info.Version))
	if !isGoodVersion(info.Version) {
		const format = "python >= %d.%d required, found %q"
		return fmt.Errorf(format, minPyVersion.Major, minPyVersion.Minor, info.VersionString)
	}
	lm := &localRunnerManager{
		logger:           log,
		runnerIDToRunner: map[string]*LocalPython{},
		mu:               new(sync.Mutex),
		workerAddress:    cfg.WorkerAddress,
		cfg:              cfg,
	}

	// If user supplies which Python to use, we use it "as-is" without creating venv
	if !isUserPy {
		if !cfg.LazyLoadVEnv {
			if err := ensureVEnv(log, pyExe); err != nil {
				return fmt.Errorf("create venv: %w", err)
			}
		}
		lm.pyExe = venvPy
	} else {
		lm.pyExe = pyExe
	}
	log.Info("using python", zap.String("exe", lm.pyExe))

	configuredRunnerType = runnerTypeLocal
	runnerManager = lm
	return nil
}

func (l *localRunnerManager) Start(ctx context.Context, buildArtifacts []byte, vars map[string]string) (string, pb.RunnerClient, error) {
	r := &LocalPython{
		log: l.logger,
	}

	if l.cfg.LazyLoadVEnv {
		if err := ensureVEnv(l.logger, l.pyExe); err != nil {
			return "", nil, fmt.Errorf("create venv: %w", err)
		}
	}

	if err := r.Start(l.pyExe, buildArtifacts, vars, l.workerAddress); err != nil {
		return "", nil, err
	}

	runnerAddr := fmt.Sprintf("localhost:%d", r.port)
	client, err := dialRunner(runnerAddr)
	if err != nil {
		if err := r.Close(); err != nil {
			l.logger.Warn("close runner", zap.Error(err))
		}
		return "", nil, err
	}

	l.mu.Lock()
	l.runnerIDToRunner[r.id.String()] = r
	l.mu.Unlock()
	return r.id.String(), client, nil
}

func (*localRunnerManager) RunnerHealth(ctx context.Context, runnerID string) error { return nil }

func (l *localRunnerManager) Stop(ctx context.Context, runnerID string) error {
	l.mu.Lock()
	runner, ok := l.runnerIDToRunner[runnerID]
	l.mu.Unlock()

	if !ok {
		return errors.New("not found")
	}

	err := runner.Close()

	l.mu.Lock()
	delete(l.runnerIDToRunner, runnerID)
	l.mu.Unlock()

	return err
}

func (*localRunnerManager) Health(ctx context.Context) error { return nil }

const exeEnvKey = "AK_WORKER_PYTHON"

func isFile(path string) bool {
	finfo, err := os.Stat(path)
	if err != nil {
		return false
	}

	return !finfo.IsDir()
}

func pythonToRun(log *zap.Logger) (string, bool, error) {
	pyExe := os.Getenv(exeEnvKey)
	if pyExe != "" {
		log.Info("python from env", zap.String("python", pyExe))
		if !isFile(pyExe) {
			log.Error("can't python from env: %q", zap.String("path", pyExe))
			return "", false, fmt.Errorf("%q: not a file", pyExe)
		}

		return pyExe, true, nil
	}

	pyExe, err := findPython()
	if err != nil {
		return "", false, err
	}

	return pyExe, false, nil
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

var minPyVersion = Version{
	Major: 3,
	Minor: 11,
}

func isGoodVersion(v Version) bool {
	if v.Major < minPyVersion.Major {
		return false
	}

	return v.Minor >= minPyVersion.Minor
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.IsDir()
}

func ensureVEnv(log *zap.Logger, pyExe string) error {
	if dirExists(venvPath) {
		return nil
	}

	log.Info("creating venv", zap.String("path", venvPath))
	return createVEnv(pyExe, venvPath)
}
