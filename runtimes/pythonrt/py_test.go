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
	_, err := exec.LookPath("python")
	if err != nil {
		t.Skip("no python installed")
	}
}

func Test_createVEnv(t *testing.T) {
	if testing.Short() {
		t.Skip("short mode")
	}

	info, err := pyExeInfo(context.Background())
	if errors.Is(errors.Unwrap(err), exec.ErrNotFound) {
		t.Skip("python not found")
	}
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

	ri, err := runPython(log, "python", tarData, "simple.py:greet", env)
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

var pyVersionCases = []struct {
	version string
	major   int
	minor   int
	err     bool
}{
	{"Python 3.12.2", 3, 12, false},
	{"Python 2.7.18", 2, 7, false},
	{"Python 3", 0, 0, true},
	{"", 0, 0, true},
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
