package nodejsrt

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func TestMain(m *testing.M) {
	code := m.Run()
	runnerManager = nil
	configuredRunnerType = runnerTypeNotConfigured
	os.Exit(code)
}

func validateTar(t *testing.T, tarData []byte, fsys fs.FS) {
	inTar := make(map[string]bool)

	r := tar.NewReader(bytes.NewReader(tarData))
	for {
		hdr, err := r.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		require.NoError(t, err, "iterate tar")
		// Skip directories in tar
		if hdr.Typeflag == tar.TypeDir {
			continue
		}
		require.Truef(t, isFSFile(fsys, hdr.Name), "%q - not on fs", hdr.Name)
		inTar[hdr.Name] = true
	}

	entries, err := fs.ReadDir(fsys, ".")
	require.NoError(t, err, "read dir")
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		require.Containsf(t, inTar, name, "%q not in tar", name)
	}
}

func isFSFile(fsys fs.FS, path string) bool {
	info, err := fs.Stat(fsys, path)
	if err != nil {
		return false
	}

	return !info.IsDir()
}

func newSVC(t *testing.T) *nodejsSvc {
	rt, err := newSvc(Configs.Default, zap.NewNop())
	require.NoError(t, err, "New")
	svc, ok := rt.(*nodejsSvc)
	require.Truef(t, ok, "type assertion failed, got %T", rt)

	return svc
}

func TestConfigureLocalRuntime(t *testing.T) {
	l := kittehs.Must1(zap.NewDevelopment())

	// Test with no worker address and no provider
	cfg := &Config{
		RunnerType: "local",
	}
	_, err := New(cfg, l, nil)
	require.Error(t, err, "should fail on not set worker address")

	// Test with worker address
	cfg = &Config{
		RunnerType:    "local",
		WorkerAddress: "0.0.0.0:123",
	}
	_, err = New(cfg, l, nil)
	require.NoError(t, err, "should succeed to configure")

	// Test with worker address provider
	cfg = &Config{
		RunnerType: "local",
	}
	_, err = New(cfg, l, func() string { return "0.0.0.0:123" })
	require.NoError(t, err, "should succeed to configure")
}

func Test_nodeSvc_Get(t *testing.T) {
	svc := newSVC(t)
	rt := svc.Get()
	require.NotNil(t, rt)
}

func Test_nodeSvc_Build(t *testing.T) {
	svc := newSVC(t)

	rootPath := "testdata/simple_test/"
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

func newValues(t *testing.T, runID sdktypes.RunID) (sdktypes.ModuleFunction, map[string]sdktypes.Value) {
	xid := sdktypes.NewExecutorID(runID)
	var mod sdktypes.ModuleFunction
	syscall, err := sdktypes.NewFunctionValue(xid, "syscall", nil, nil, mod)
	require.NoError(t, err)

	ak, err := sdktypes.NewStructValue(
		sdktypes.NewStringValue("ak"),
		map[string]sdktypes.Value{"syscall": syscall},
	)
	require.NoError(t, err)

	return mod, map[string]sdktypes.Value{"ak": ak}
}

func newCallbacks(svc *nodejsSvc) *sdkservices.RunCallbacks {
	cbs := sdkservices.RunCallbacks{
		Call: func(
			ctx context.Context,
			rid sdktypes.RunID,
			v sdktypes.Value,
			args []sdktypes.Value,
			kwargs map[string]sdktypes.Value,
		) (sdktypes.Value, error) {
			return svc.Call(ctx, v, args, kwargs)
		},
		Load: func(ctx context.Context, rid sdktypes.RunID, path string) (map[string]sdktypes.Value, error) {
			return map[string]sdktypes.Value{}, nil
		},
		Print: func(ctx context.Context, rid sdktypes.RunID, text string) error {
			return nil
		},
	}

	return &cbs
}

func setupServer(l *zap.Logger) (net.Listener, error) {
	mux := &http.ServeMux{}
	ConfigureWorkerGRPCHandler(l, mux)
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}

	// Create config for local runner
	cfg := &Config{
		RunnerType:    "local",
		LogRunnerCode: true,
	}

	// Create NodeJS runtime which will configure the local runner manager
	_, err = New(cfg, l, func() string {
		port := listener.Addr().(*net.TCPAddr).Port
		return fmt.Sprintf("localhost:%d", port)
	})
	if err != nil {
		return nil, fmt.Errorf("configure local runner manager: %w", err)
	}

	errChan := make(chan error)
	go func() {
		if err := http.Serve(listener, h2c.NewHandler(mux, &http2.Server{})); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
			return
		}
	}()

	select {
	case err := <-errChan:
		return nil, err
	case <-time.After(2 * time.Second): // give server time to start
		return listener, nil
	}
}

func Test_nodeSvc_Run(t *testing.T) {
	svc := newSVC(t)
	require.NotNil(t, svc.log, "nil logger")

	fsys := os.DirFS("testdata/simple_test")
	tarData, err := createTar(fsys)
	require.NoError(t, err, "create tar")

	ctx, cancel := testCtx(t)
	defer cancel()
	runID := sdktypes.NewRunID()
	mainPath := "index.ts"
	compiled := map[string][]byte{
		archiveKey: tarData,
	}

	cbs := newCallbacks(svc)
	mod, values := newValues(t, runID)
	xid := sdktypes.NewExecutorID(runID)

	server, err := setupServer(svc.log)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := server.Close(); err != nil {
			t.Log(err)
		}
	}()

	sessionID := sdktypes.NewSessionID()
	run, err := svc.Run(ctx, runID, sessionID, mainPath, compiled, values, cbs)
	require.NoError(t, err, "run")

	fn, err := sdktypes.NewFunctionValue(xid, "hello", nil, nil, mod)
	require.NoError(t, err, "new function")

	_, err = run.Call(ctx, fn, nil, nil)
	require.NoError(t, err, "call")
}

func TestProgramError(t *testing.T) {
	// Create a simple JS file that will throw an error
	jsCode := []byte(`
		function handle() {
			throw new Error("test error");
		}
		module.exports = { handle };
	`)

	jsFile := "error.js"
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	hdr := &tar.Header{
		Name: jsFile,
		Size: int64(len(jsCode)),
		Mode: 0o644,
	}
	err := tw.WriteHeader(hdr)
	require.NoError(t, err)
	_, err = tw.Write(jsCode)
	tw.Close()

	compiled := map[string][]byte{
		archiveKey: buf.Bytes(),
	}

	require.NoError(t, err)
	svc := newSVC(t)
	server, err := setupServer(svc.log)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = server.Close()
	}()

	runID := sdktypes.NewRunID()
	mod, values := newValues(t, runID)
	cbs := newCallbacks(svc)
	ctx, cancel := testCtx(t)
	defer cancel()
	sid := sdktypes.NewSessionID()
	_, err = svc.Run(ctx, runID, sid, jsFile, compiled, values, cbs)
	require.NoError(t, err)

	xid := sdktypes.NewExecutorID(runID)
	fn, err := sdktypes.NewFunctionValue(xid, "handle", nil, nil, mod)
	require.NoError(t, err, "new function")

	_, err = svc.Call(ctx, fn, nil, nil)
	require.Error(t, err)
	t.Logf("ERROR:\n%s", err)
	require.Contains(t, err.Error(), "test error")
	require.Contains(t, err.Error(), "error.js")
}
