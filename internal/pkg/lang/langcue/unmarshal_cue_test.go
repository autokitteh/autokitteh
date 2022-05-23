//go:build unit

package langcue

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalCueData(t *testing.T) {
	var dst map[string]interface{}

	require.NoError(
		t,
		UnmarshalCue(context.Background(), []byte(`package main

import "encoding/json"
import "autokitteh.io/values"

_values: values.#Values & {
	cat: "meow"
}

v: 42
test: values.#Value & json.Marshal(v)
sound: _values.cat
		`), &dst),
	)

	assert.EqualValues(
		t,
		map[string]interface{}{
			"v":     42,
			"test":  "42",
			"sound": "meow",
		}, dst)
}
