package pythonrt

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"testing"
	"time"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.uber.org/zap"

	"github.com/stretchr/testify/require"
)

func validateTar(t *testing.T, tarData []byte, fsys fs.FS) {
	inTar := make(map[string]bool)

	r := tar.NewReader(bytes.NewReader(tarData))
	for {
		hdr, err := r.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		require.NoError(t, err, "iterate tar")
		require.Truef(t, isFSFile(fsys, hdr.Name), "%q - not on fs", hdr.Name)
		inTar[hdr.Name] = true
	}

	entries, err := fs.ReadDir(fsys, ".")
	require.NoError(t, err, "read dir")
	for _, e := range entries {
		require.Truef(t, inTar[e.Name()], "%q not in tar", e.Name())
	}
}

func isFSFile(fsys fs.FS, path string) bool {
	info, err := fs.Stat(fsys, path)
	if err != nil {
		return false
	}

	return !info.IsDir()
}

func newSVC(t *testing.T) pySvc {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err, "create logger")

	return pySvc{
		log: logger,
	}
}

func Test_pySvc_Get(t *testing.T) {
	svc := newSVC(t)
	rt := svc.Get()
	require.NotNil(t, rt)
}

func Test_pySvc_Build(t *testing.T) {
	svc := newSVC(t)

	rootPath := "testdata/simple/"
	fsys := os.DirFS(rootPath)

	ctx, cancel := testCtx(t)
	defer cancel()
	art, err := svc.Build(ctx, fsys, ".", nil)
	require.NoError(t, err)

	p := art.ToProto()
	data := p.CompiledData[archiveKey]
	validateTar(t, data, fsys)
}

func testCtx(t *testing.T) (context.Context, context.CancelFunc) {
	d, ok := t.Deadline()
	if !ok {
		d = time.Now().Add(3 * time.Second)
	}

	return context.WithDeadline(context.Background(), d)
}

func Test_pySvc_Run(t *testing.T) {
	skipIfNoPython(t)

	rt, err := New()
	require.NoError(t, err, "New")
	svc, ok := rt.(*pySvc)
	require.True(t, ok, "type assertion failed")
	require.NotNil(t, svc.log, "nil logger")

	fsys := os.DirFS("testdata/simple")
	tarData, err := createTar(fsys)
	require.NoError(t, err, "create tar")

	ctx, cancel := testCtx(t)
	defer cancel()
	runID := sdktypes.NewRunID()
	mainPath := "simple.py:greet"
	compiled := map[string][]byte{
		archiveKey: tarData,
	}

	cbs := sdkservices.RunCallbacks{
		Call: func(
			ctx context.Context,
			rid sdktypes.RunID,
			v sdktypes.Value,
			args []sdktypes.Value,
			kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
			return svc.Call(ctx, v, args, kwargs)
		},
		Load: func(ctx context.Context, rid sdktypes.RunID, path string) (map[string]sdktypes.Value, error) {
			return map[string]sdktypes.Value{}, nil
		},
		Print: func(ctx context.Context, rid sdktypes.RunID, msg string) {},
	}

	xid := sdktypes.NewExecutorID(sdktypes.NewRunID())

	var mod sdktypes.ModuleFunction
	syscall, err := sdktypes.NewFunctionValue(xid, "syscall", nil, nil, mod)
	require.NoError(t, err)

	ak, err := sdktypes.NewStructValue(
		sdktypes.NewStringValue("ak"),
		map[string]sdktypes.Value{"syscall": syscall},
	)
	require.NoError(t, err)
	values := map[string]sdktypes.Value{"ak": ak}
	run, err := svc.Run(ctx, runID, mainPath, compiled, values, &cbs)
	require.NoError(t, err, "run")

	fn, err := sdktypes.NewFunctionValue(xid, "greet", nil, nil, mod)
	require.NoError(t, err, "new function")

	kwargs := map[string]sdktypes.Value{
		"event_id": sdktypes.NewStringValue("007"),
		"data": sdktypes.NewDictValueFromStringMap(map[string]sdktypes.Value{
			"body": sdktypes.NewBytesValue([]byte(`{"user": "joe", "id": 7}`)),
		}),
	}
	_, err = run.Call(ctx, fn, nil, kwargs)
	require.NoError(t, err, "call")
}

var isGoodVersionCasess = []struct {
	version Version
	ok      bool
}{
	{Version{2, 6}, false},
	{Version{3, minPyVersion.Minor - 1}, false},
	{Version{3, minPyVersion.Minor}, true},
	{Version{3, minPyVersion.Minor + 1}, true},
}

func Test_isGoodVersion(t *testing.T) {
	for _, tc := range isGoodVersionCasess {
		name := fmt.Sprintf("%d.%d", tc.version.Major, tc.version.Minor)
		t.Run(name, func(t *testing.T) {
			ok := isGoodVersion(tc.version)
			require.Equal(t, tc.ok, ok)
		})
	}
}

const exeCodeTemplate = `#!/bin/bash

echo Python %d.%d.7
`

func TestNewBadVersion(t *testing.T) {
	dirName := t.TempDir()
	exe := path.Join(dirName, "python")
	file, err := os.Create(exe)
	require.NoError(t, err)

	exeCode := fmt.Sprintf(exeCodeTemplate, minPyVersion.Major, minPyVersion.Minor-1)
	_, err = file.Write([]byte(exeCode))
	require.NoError(t, err)
	file.Close()

	err = os.Chmod(exe, 0755)
	require.NoError(t, err)
	t.Setenv("PATH", dirName)

	_, err = New()
	require.Error(t, err)
}
