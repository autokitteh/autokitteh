package pythonrt

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
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
		require.Truef(t, isFile(fsys, hdr.Name), "%q - not on fs", hdr.Name)
		inTar[hdr.Name] = true
	}

	entries, err := fs.ReadDir(fsys, ".")
	require.NoError(t, err, "read dir")
	for _, e := range entries {
		require.Truef(t, inTar[e.Name()], "%q not in tar", e.Name())
	}
}

func isFile(fsys fs.FS, path string) bool {
	info, err := fs.Stat(fsys, path)
	if err != nil {
		return false
	}

	return !info.IsDir()
}

func newSVC(t *testing.T) pySVC {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err, "create logger")

	return pySVC{
		log: logger,
	}
}

func Test_pySVC_Get(t *testing.T) {
	svc := newSVC(t)
	rt := svc.Get()
	require.NotNil(t, rt)
}

// func (p pySVC) Build(ctx context.Context, fs fs.FS, path string, values []sdktypes.Symbol) (sdktypes.BuildArtifact, error) {
func Test_pySVC_Build(t *testing.T) {
	svc := newSVC(t)

	rootPath := "testdata/review/"
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

func Test_pySVC_Run(t *testing.T) {
	rt, err := New()
	require.NoError(t, err, "New")
	svc, ok := rt.(*pySVC)
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
	}

	run, err := svc.Run(ctx, runID, mainPath, compiled, nil, &cbs)
	require.NoError(t, err, "run")

	xid := sdktypes.NewExecutorID(sdktypes.NewRunID())
	var modFn sdktypes.ModuleFunction
	fn, err := sdktypes.NewFunctionValue(xid, "greet", nil, nil, modFn)
	require.NoError(t, err, "new function")

	kwargs := map[string]sdktypes.Value{
		"event_id": sdktypes.NewStringValue("007"),
	}
	_, err = run.Call(ctx, fn, nil, kwargs)
	require.NoError(t, err, "call")
}
