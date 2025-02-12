package nodejsrt

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

//go:embed testdata/simple.tar
var tarData []byte

func TestRunner_Start(t *testing.T) {
	log := zap.NewExample()
	defer log.Sync() //nolint:all

	envKey := "AK_TEST_ENV"
	t.Setenv(envKey, "A")
	env := map[string]string{
		envKey: "B",
	}

	r := LocalNodeJS{
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
	cmd := exec.Command("ps", "e", "-ww", "-p", strconv.Itoa(pid))
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
// 	r := LocalNodeJS{
// 		log: log,
// 	}
// 	err = r.Start("python", tarData, nil, workerAddr)
// 	require.NoError(t, err)

// 	defer r.Close() //nolint:all

// 	client, err := dialRunner(fmt.Sprintf("localhost:%d", r.port))
// 	require.NoError(t, err)
// 	req := pb.ExportsRequest{
// 		FileName: "simple.js",
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
	r := LocalNodeJS{
		log: log,
	}

	err := r.Start("python", tarData, nil, "")
	require.NoError(t, err)

	r.Close()

	require.NoDirExists(t, r.runnerDir)
	require.NoDirExists(t, r.userDir)
}
