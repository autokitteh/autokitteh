package sessionsgrpcsvc

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func TestUnpackJSONObject(t *testing.T) {
	obj := map[string]any{
		"key": "value",
		"key2": map[string]any{
			"key3": "value3",
		},
		"key4": []any{
			"item1",
			"item2",
		},
	}

	objstr := string(kittehs.Must1(json.Marshal(obj)))

	out := make(map[string]*sdktypes.ValuePB)

	require.NoError(t, unpackJSONObject(objstr, out))

	fs, err := kittehs.TransformMapValuesError(out, sdktypes.ValueFromProto)
	require.NoError(t, err)

	v := sdktypes.NewDictValueFromStringMap(fs)

	w := sdktypes.ValueWrapper{SafeForJSON: true}

	u, err := w.Unwrap(v)
	require.NoError(t, err)
	require.Equal(t, obj, u)
}
