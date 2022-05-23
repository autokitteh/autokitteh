package langcue

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/autokitteh/L"
	"go.autokitteh.dev/sdk/api/apiprogram"
	"go.autokitteh.dev/sdk/api/apivalues"
)

func TestJSONData(t *testing.T) {
	l, err := NewJSONDataLang(L.Nop, "json-data")
	require.NoError(t, err)

	mod, err := l.CompileModule(
		context.Background(),
		apiprogram.MustParsePathString("fs:test.json"),
		[]byte(`{"cat": "meow"}`),
		nil,
	)
	require.NoError(t, err)

	vs, _, err := l.RunModule(context.Background(), nil, mod)
	require.NoError(t, err)
	assert.EqualValues(
		t,
		map[string]interface{}{"cat": "meow"},
		apivalues.UnwrapValuesMap(vs),
	)
}

func TestJSONProg(t *testing.T) {
	l, err := NewJSONProgLang(L.Nop, "json-prog")
	require.NoError(t, err)

	mod, err := l.CompileModule(
		context.Background(),
		apiprogram.MustParsePathString("fs:test.kitteh.json"),
		[]byte(`{"consts": {"cat": "meow"}}`),
		nil,
	)
	require.NoError(t, err)

	vs, _, err := l.RunModule(context.Background(), nil, mod)
	require.NoError(t, err)
	assert.EqualValues(
		t,
		map[string]interface{}{"cat": "meow"},
		apivalues.UnwrapValuesMap(vs),
	)
}

func TestYAMLData(t *testing.T) {
	l, err := NewYAMLDataLang(L.Nop, "yaml-data")
	require.NoError(t, err)

	mod, err := l.CompileModule(
		context.Background(),
		apiprogram.MustParsePathString("fs:test.yaml"),
		[]byte(`cat: meow`),
		nil,
	)
	require.NoError(t, err)

	vs, _, err := l.RunModule(context.Background(), nil, mod)
	require.NoError(t, err)
	assert.EqualValues(
		t,
		map[string]interface{}{"cat": "meow"},
		apivalues.UnwrapValuesMap(vs),
	)
}

func TestYAMLProg(t *testing.T) {
	l, err := NewYAMLProgLang(L.Nop, "json-prog")
	require.NoError(t, err)

	mod, err := l.CompileModule(
		context.Background(),
		apiprogram.MustParsePathString("fs:test.kitteh.json"),
		[]byte("consts:\n  cat: meow"),
		nil,
	)
	require.NoError(t, err)

	vs, _, err := l.RunModule(context.Background(), nil, mod)
	require.NoError(t, err)
	assert.EqualValues(
		t,
		map[string]interface{}{"cat": "meow"},
		apivalues.UnwrapValuesMap(vs),
	)
}

func TestCueData(t *testing.T) {
	l, err := NewCueDataLang(L.Nop, "cue-data")
	require.NoError(t, err)

	mod, err := l.CompileModule(
		context.Background(),
		apiprogram.MustParsePathString("fs:test.kitteh.json"),
		[]byte(`cat: "meow"`),
		nil,
	)
	require.NoError(t, err)

	vs, _, err := l.RunModule(context.Background(), nil, mod)
	require.NoError(t, err)
	assert.EqualValues(
		t,
		map[string]interface{}{"cat": "meow"},
		apivalues.UnwrapValuesMap(vs),
	)
}

func TestCueProg(t *testing.T) {
	l, err := NewCueProgLang(L.Nop, "cue-prog")
	require.NoError(t, err)

	mod, err := l.CompileModule(
		context.Background(),
		apiprogram.MustParsePathString("fs:test.kitteh.json"),
		[]byte(`consts: cat: "meow"`),
		nil,
	)
	require.NoError(t, err)

	vs, _, err := l.RunModule(context.Background(), nil, mod)
	require.NoError(t, err)
	assert.EqualValues(
		t,
		map[string]interface{}{"cat": "meow"},
		apivalues.UnwrapValuesMap(vs),
	)
}
