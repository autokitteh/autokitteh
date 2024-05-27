package pythonrt

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func skipIfNoPython(t *testing.T) {
	_, err := pyExeInfo(context.Background())
	if errors.Is(err, exec.ErrNotFound) {
		t.Skip("no python installed")
	}

	if err != nil {
		t.Logf("error getting Python info: %s", err)
	}
}

func Test_createVEnv(t *testing.T) {
	skipIfNoPython(t)

	if testing.Short() {
		t.Skip("short mode")
	}

	info, err := pyExeInfo(context.Background())
	require.NoError(t, err)

	venvPath := path.Join(t.TempDir(), "venv")
	err = createVEnv(info.Exe, venvPath)
	require.NoError(t, err)
}

//go:embed testdata/simple.tar
var tarData []byte

func Test_runPython(t *testing.T) {
	skipIfNoPython(t)

	log := zap.NewExample()
	defer log.Sync() //nolint:all

	envKey := "AK_TEST_ENV"
	t.Setenv(envKey, "A")
	env := map[string]string{
		envKey: "B",
	}

	opts := runOptions{
		log:      log,
		pyExe:    "python",
		tarData:  tarData,
		rootPath: "simple.py:greet",
		env:      env,
		stdout:   os.Stdout,
		stderr:   os.Stderr,
	}
	ri, err := runPython(opts)
	require.NoError(t, err)
	defer ri.proc.Kill() //nolint:all

	procEnv := processEnv(t, ri.proc.Pid)
	require.Equal(t, env[envKey], procEnv[envKey], "env override")
}

var envRe = regexp.MustCompile(`([^ ]+)=([^ \n]+)`)

func processEnv(t *testing.T, pid int) map[string]string {
	var buf bytes.Buffer
	cmd := exec.Command("ps", "e", "-ww", "-p", fmt.Sprintf("%d", pid))
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr

	require.NoError(t, cmd.Run())

	env := make(map[string]string)
	for _, match := range envRe.FindAllStringSubmatch(buf.String(), -1) {
		env[match[1]] = match[2]
	}
	return env
}

func genExe(t *testing.T, name string) {
	file, err := os.Create(name)
	require.NoError(t, err)
	file.Close()

	// We must set executable bit on the file otherwise exec.LookPath will ignore it.
	err = os.Chmod(name, 0o766)
	require.NoError(t, err)
}

func Test_findPython(t *testing.T) {
	dirName := t.TempDir()
	t.Setenv("PATH", dirName)

	// No Python
	_, err := findPython()
	require.Error(t, err)

	// python
	pyExe := path.Join(dirName, "python")
	genExe(t, pyExe)
	out, err := findPython()
	require.NoError(t, err)
	require.Equal(t, pyExe, out)

	// python & python3, should be python3
	py3Exe := path.Join(dirName, "python3")
	genExe(t, py3Exe)
	out, err = findPython()
	require.NoError(t, err)
	require.Equal(t, py3Exe, out)

	// Symlink
	for _, name := range []string{pyExe, py3Exe} {
		err = os.Remove(name)
		require.NoError(t, err)
	}

	exe := path.Join(dirName, "python3.12")
	genExe(t, exe)
	link := path.Join(dirName, "python3")
	err = os.Symlink(exe, link)
	require.NoError(t, err)
	out, err = findPython()
	require.NoError(t, err)
	require.Equal(t, link, out)

	// Environment
	envDir := t.TempDir()
	envExe := path.Join(envDir, "pypy3")
	t.Setenv(exeEnvKey, envExe)
	_, err = findPython()
	require.Error(t, err)

	genExe(t, envExe)
	out, err = findPython()
	require.NoError(t, err)
	require.Equal(t, envExe, out)
}

var pyVersionCases = []struct {
	version string
	major   int
	minor   int
	err     bool
}{
	{"Python 3.12.2", 3, 12, false},
	{"Python 2.7.18", 2, 7, false},
	{"Python 3", 0, 0, true},
	{"Python", 0, 0, true},
	{"sl 1.2.3", 0, 0, true},
	{"", 0, 0, true},
	{"Python 3.10.13 (fc59e61cfbff, Jan 17 2024, 05:35:45)", 3, 10, false},
}

func Test_parsePyVersion(t *testing.T) {
	for _, tc := range pyVersionCases {
		major, minor, err := parsePyVersion(tc.version)
		if tc.err {
			require.Error(t, err)
			continue
		}

		require.NoError(t, err)
		require.Equal(t, tc.major, major)
		require.Equal(t, tc.minor, minor)
	}
}
