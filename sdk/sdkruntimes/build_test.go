package sdkruntimes

import (
	"context"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type fakeRuntime struct {
	d     sdktypes.Runtime
	paths []string
	add   func(*fakeRuntime)
}

func (f *fakeRuntime) Get() sdktypes.Runtime { return f.d }

func (f *fakeRuntime) Build(_ context.Context, _ fs.FS, path string, _ []sdktypes.Symbol) (sdktypes.BuildArtifact, error) {
	f.paths = append(f.paths, path)
	return sdktypes.InvalidBuildArtifact, nil
}

func (*fakeRuntime) Run(context.Context, sdktypes.RunID, sdktypes.SessionID, string, map[string][]byte, map[string]sdktypes.Value, *sdkservices.RunCallbacks) (sdkservices.Run, error) {
	return nil, nil
}

func (f *fakeRuntime) New() (sdkservices.Runtime, error) {
	f.add(f)
	return f, nil
}

func newRuntime(add func(*fakeRuntime), name string, filewise bool) *Runtime {
	d := kittehs.Must1(sdktypes.RuntimeFromProto(&sdktypes.RuntimePB{
		Name:           name,
		FileExtensions: []string{name},
		FilewiseBuild:  filewise,
	}))

	return &Runtime{
		Desc: d,
		New:  (&fakeRuntime{d: d, add: add}).New,
	}
}

func NewRuntimes() (sdkservices.Runtimes, map[string]*fakeRuntime) {
	created := make(map[string]*fakeRuntime)
	add := func(f *fakeRuntime) {
		created[f.d.Name().String()] = f
	}

	return kittehs.Must1(New([]*Runtime{newRuntime(add, "a", true), newRuntime(add, "b", false)})), created
}

func TestBuild(t *testing.T) {
	mfs := kittehs.Must1(kittehs.MapToMemFS(map[string][]byte{
		"a.a": []byte("a.a"),
		"b.a": []byte("b.a"),
		"c.b": []byte("c.b"),
		"d.b": []byte("d.b"),
	}))

	rts, built := NewRuntimes()

	_, err := Build(context.Background(), rts, mfs, nil, nil)
	if assert.NoError(t, err) {
		assert.Equal(t, []string{"a.a", "b.a"}, built["a"].paths)
		assert.Equal(t, []string{"."}, built["b"].paths)
	}
}

func TestBuildSubdirs(t *testing.T) {
	mfs := kittehs.Must1(kittehs.MapToMemFS(map[string][]byte{
		"a.a":     []byte("a.a"),
		"foo/b.a": []byte("b.a"),
		"c.b":     []byte("c.b"),
		"bar/d.b": []byte("d.b"),
	}))

	rts, built := NewRuntimes()

	_, err := Build(context.Background(), rts, mfs, nil, nil)
	if assert.NoError(t, err) {
		assert.Equal(t, []string{"a.a", "foo/b.a"}, built["a"].paths)
		assert.Equal(t, []string{"."}, built["b"].paths)
	}
}

func TestBuildMixed(t *testing.T) {
	mfs := kittehs.Must1(kittehs.MapToMemFS(map[string][]byte{
		"a.a":     []byte("a.a"),
		"foo/b.a": []byte("b.a"),
		"c.b":     []byte("c.b"),
		"foo/d.b": []byte("d.b"),
	}))

	rts, built := NewRuntimes()

	_, err := Build(context.Background(), rts, mfs, nil, nil)
	if assert.NoError(t, err) {
		assert.Equal(t, []string{"a.a", "foo/b.a"}, built["a"].paths)
		assert.Equal(t, []string{"."}, built["b"].paths)
	}
}
