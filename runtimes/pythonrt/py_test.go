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

func Test_createVEnv(t *testing.T) {
	info, err := pyExecInfo(context.Background())
	if errors.Is(errors.Unwrap(err), exec.ErrNotFound) {
		t.Skipf("python not found")
	}
	require.NoError(t, err)

	venvPath := path.Join(t.TempDir(), "venv")
	err = createVEnv(info.Exe, venvPath)
	require.NoError(t, err)
}

var (
	//go:embed testdata/simple.tar
	tarData []byte
	envRe   = regexp.MustCompile(`([A-Z]+)=([^ ]+)`) //nolint:all (see TODO below)
)

// func runPython(log *zap.Logger, tarData []byte, rootPath string, env map[string]string) (*pyRunInfo, error) {
func Test_runPython(t *testing.T) {
	log := zap.NewExample()
	defer log.Sync() //nolint:all

	envKey := "AK_TEST_ENV"
	t.Setenv(envKey, "A")
	env := map[string]string{
		envKey: "B",
	}

	ri, err := runPython(log, tarData, "simple.py:greet", env)
	require.NoError(t, err)
	defer ri.proc.Kill() //nolint:all

	/* TODO: There's a buf in processEnv
	procEnv := processEnv(t, ri.proc.Pid)
	require.Equal(t, env[envKey], procEnv[envKey], "env override")
	*/
}

func processEnv(t *testing.T, pid int) map[string]string { //nolint:all
	var buf bytes.Buffer
	cmd := exec.Command("ps", "e", "-ww", "-p", fmt.Sprintf("%d", pid))
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr

	require.NoError(t, cmd.Run())
	env := make(map[string]string)
	for _, match := range envRe.FindAllStringSubmatch(buf.String(), -1) {
		t.Logf("ENV: %q -> %q", match[0], match[1])
		env[match[0]] = match[1]
	}
	return env
}
