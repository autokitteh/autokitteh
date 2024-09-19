package pythonrt

import (
	"bytes"
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
	pyExe, err := findPython()
	if err != nil {
		t.Skip("no python installed")
	}

	ctx, cancel := testCtx(t)
	defer cancel()

	_, err = pyExeInfo(ctx, pyExe)
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

	pyExe, err := findPython()
	require.NoError(t, err)

	ctx, cancel := testCtx(t)
	defer cancel()

	info, err := pyExeInfo(ctx, pyExe)
	require.NoError(t, err)

	venvPath := path.Join(t.TempDir(), "venv")
	err = createVEnv(info.Exe, venvPath)
	require.NoError(t, err)
}

//go:embed testdata/simple.tar
var tarData []byte

func TestRunner_Start(t *testing.T) {
	skipIfNoPython(t)

	log := zap.NewExample()
	defer log.Sync() //nolint:all

	envKey := "AK_TEST_ENV"
	t.Setenv(envKey, "A")
	env := map[string]string{
		envKey: "B",
	}

	r := LocalPython{
		log: log,
	}
	err := r.Start("python", tarData, env, "")
	require.NoError(t, err)

	defer r.Close() //nolint:all

	procEnv := processEnv(t, r.proc.Pid)
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

const exeCodeTemplate = `#!/bin/bash

echo Python %d.%d.7
`

func genExe(t *testing.T, path string, major, minor int) {
	file, err := os.Create(path)
	require.NoError(t, err)
	defer file.Close()

	fmt.Fprintf(file, exeCodeTemplate, major, minor)

	// We must set executable bit on the file otherwise exec.LookPath will ignore it.
	err = os.Chmod(path, 0o766)
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
	genExe(t, pyExe, minPyVersion.Major, minPyVersion.Minor)
	out, err := findPython()
	require.NoError(t, err)
	require.Equal(t, pyExe, out)

	// python & python3, should be python3
	py3Exe := path.Join(dirName, "python3")
	genExe(t, py3Exe, minPyVersion.Major, minPyVersion.Minor)
	out, err = findPython()
	require.NoError(t, err)
	require.Equal(t, py3Exe, out)

	// Symlink
	for _, name := range []string{pyExe, py3Exe} {
		err = os.Remove(name)
		require.NoError(t, err)
	}

	exe := path.Join(dirName, "python3.12")
	genExe(t, exe, minPyVersion.Major, minPyVersion.Minor)
	link := path.Join(dirName, "python3")
	err = os.Symlink(exe, link)
	require.NoError(t, err)
	out, err = findPython()
	require.NoError(t, err)
	require.Equal(t, link, out)
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

// TODO: What to here
// func Test_pyExports(t *testing.T) {
// 	skipIfNoPython(t)

// 	log := zap.NewExample()
// 	defer log.Sync() //nolint:all

// 	runID := sdktypes.NewRunID()
// 	xid := sdktypes.NewExecutorID(runID)
// 	svc := newWorkerGRPCHandler(log, nil, runID, xid, sdktypes.Nothing)
// 	err := svc.Start()
// 	require.NoError(t, err)

// 	workerAddr := fmt.Sprintf("localhost:%d", svc.port)
// 	r := LocalPython{
// 		log: log,
// 	}
// 	err = r.Start("python", tarData, nil, workerAddr)
// 	require.NoError(t, err)

// 	defer r.Close() //nolint:all

// 	client, err := dialRunner(fmt.Sprintf("localhost:%d", r.port))
// 	require.NoError(t, err)
// 	req := pb.ExportsRequest{
// 		FileName: "simple.py",
// 	}
// 	ctx, cancel := testCtx(t)
// 	defer cancel()

// 	resp, err := client.Exports(ctx, &req)
// 	require.NoError(t, err)

// 	require.Equal(t, []string{"greet"}, resp.Exports)
// }

const testRunnerPath = "/tmp/zzz/ak_runner"

var adjustCases = []struct {
	name     string
	env      []string
	expected []string
}{
	{"empty env", nil, []string{"PYTHONPATH=" + testRunnerPath}},
	{
		name:     "regular",
		env:      []string{"HOME=/home/ak", "PYTHONPATH=x"},
		expected: []string{"HOME=/home/ak", "PYTHONPATH=x:" + testRunnerPath},
	},
	{
		name:     "last",
		env:      []string{"PYTHONPATH=x", "HOME=/home/ak", "PYTHONPATH=y"},
		expected: []string{"PYTHONPATH=x", "HOME=/home/ak", "PYTHONPATH=y:" + testRunnerPath},
	},
}

func TestRunner_Close(t *testing.T) {
	log := zap.NewExample()
	defer log.Sync() //nolint:all
	r := LocalPython{
		log: log,
	}

	err := r.Start("python", tarData, nil, "")
	require.NoError(t, err)

	r.Close()

	require.NoDirExists(t, r.runnerDir)
	require.NoDirExists(t, r.userDir)
}
