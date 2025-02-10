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
	"strings"
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
		if name == "__pycache__" {
			continue
		}
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

	err := configureLocalRunnerManager(l, LocalRunnerManagerConfig{})
	require.Error(t, err, "should fail on not set worker address")

	err = configureLocalRunnerManager(l, LocalRunnerManagerConfig{WorkerAddress: "0.0.0.0:123"})
	require.NoError(t, err, "should succeed to configure")

	err = configureLocalRunnerManager(l, LocalRunnerManagerConfig{WorkerAddressProvider: func() string { return "" }})
	require.NoError(t, err, "should succeed to configure")
}

func Test_pySvc_Get(t *testing.T) {
	svc := newSVC(t)
	rt := svc.Get()
	require.NotNil(t, rt)
}

func Test_pySvc_Build(t *testing.T) {
	skipIfNoPython(t)
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
		Print: func(ctx context.Context, rid sdktypes.RunID, msg string) {},
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

	if err := configureLocalRunnerManager(l, LocalRunnerManagerConfig{
		WorkerAddressProvider: func() string {
			port := listener.Addr().(*net.TCPAddr).Port
			return fmt.Sprintf("localhost:%d", port)
		},
		LogCodeRunnerCode: true,
	}); err != nil {
		return nil, err
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

func Test_pySvc_Run(t *testing.T) {
	skipIfNoPython(t)

	svc := newSVC(t)
	require.NotNil(t, svc.log, "nil logger")

	fsys := os.DirFS("testdata/simple")
	tarData, err := createTar(fsys)
	require.NoError(t, err, "create tar")

	ctx, cancel := testCtx(t)
	defer cancel()
	runID := sdktypes.NewRunID()
	mainPath := "simple.js"
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

	fn, err := sdktypes.NewFunctionValue(xid, "greet", nil, nil, mod)
	require.NoError(t, err, "new function")

	body := sdktypes.NewDictValueFromStringMap(map[string]sdktypes.Value{
		"bytes": sdktypes.NewBytesValue([]byte(`{"user": "joe", "id": 7}`)),
	})
	kwargs := map[string]sdktypes.Value{
		"event_id": sdktypes.NewStringValue("007"),
		"data": sdktypes.NewDictValueFromStringMap(map[string]sdktypes.Value{
			"body": body,
		}),
	}
	_, err = run.Call(ctx, fn, nil, kwargs)
	require.NoError(t, err, "call")
}

func Test_pySvc_Build_PyCache(t *testing.T) {
	skipIfNoPython(t)
	svc := newSVC(t)

	rootPath := "testdata/pycache/"
	fsys := os.DirFS(rootPath)

	ctx, cancel := testCtx(t)
	defer cancel()
	art, err := svc.Build(ctx, fsys, ".", nil)
	require.NoError(t, err)

	p := art.ToProto()
	data := p.CompiledData[archiveKey]
	r := tar.NewReader(bytes.NewReader(data))

	for {
		hdr, err := r.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		require.NoError(t, err, "iterate tar")
		require.Falsef(t, strings.Contains(hdr.Name, "__pycache__"), "%q", hdr.Name)
	}
}

var progErrCode = []byte(`
def handle(event):
    1 / 0
`)

func TestProgramError(t *testing.T) {
	skipIfNoPython(t)

	// configureLocalRunnerManager()
	pyFile := "progerr.py"
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	hdr := &tar.Header{
		Name: pyFile,
		Size: int64(len(progErrCode)),
		Mode: 0o644,
	}
	err := tw.WriteHeader(hdr)
	require.NoError(t, err)
	_, err = tw.Write(progErrCode)
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
	_, err = svc.Run(ctx, runID, sid, pyFile, compiled, values, cbs)
	require.NoError(t, err)

	xid := sdktypes.NewExecutorID(runID)
	fn, err := sdktypes.NewFunctionValue(xid, "handle", nil, nil, mod)
	require.NoError(t, err, "new function")

	kwargs := map[string]sdktypes.Value{}

	_, err = svc.Call(ctx, fn, nil, kwargs)
	require.Error(t, err)
	t.Logf("ERROR:\n%s", err)
	// There no way to check that err is a ProgramError since it's wrapped by unexported programError
	require.Contains(t, err.Error(), "division by zero")
	require.Contains(t, err.Error(), "progerr.py")
}
