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

func genExe(t *testing.T, name string) {
	file, err := os.Create(name)
	require.NoError(t, err)
	file.Close()
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
	pyExe = path.Join(dirName, "python3")
	genExe(t, pyExe)
	out, err = findPython()
	require.NoError(t, err)
	require.Equal(t, pyExe, out)
}
