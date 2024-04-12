package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBodyToStructJSON(t *testing.T) {
	v := bodyToStruct([]byte(`{"i": 1, "f": 1.2}`), nil)
	require.True(t, v.IsStruct())

	json := v.GetStruct().Fields()["json"]
	require.True(t, json.IsFunction())

	data, err := json.GetFunction().ConstValue()
	require.NoError(t, err)

	require.True(t, data.IsDict())

	fs, err := data.GetDict().ToStringValuesMap()
	require.NoError(t, err)

	i := fs["i"]
	if assert.True(t, i.IsInteger()) {
		assert.Equal(t, int64(1), i.GetInteger().Value())
	}

	f := fs["f"]
	if assert.True(t, f.IsFloat()) {
		assert.Equal(t, 1.2, f.GetFloat().Value())
	}
}
