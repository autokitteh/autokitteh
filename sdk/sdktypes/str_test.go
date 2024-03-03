package sdktypes_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func TestStr(t *testing.T) {
	var n sdktypes.Symbol

	n, err := sdktypes.ParseSymbol("test")
	if assert.NoError(t, err) {
		assert.Equal(t, "test", n.String())
	}

	assert.Equal(t, "e1b996a9b8f4e2ca87565411d377817dcecc4b7114ad4b0b6dcc42093b592c74", n.Hash())

	bs, err := json.Marshal(n)
	if assert.NoError(t, err) {
		assert.Equal(t, `"test"`, string(bs))
	}

	var nn sdktypes.Symbol
	if assert.NoError(t, nn.UnmarshalJSON([]byte(`"test"`))) {
		assert.Equal(t, "test", nn.String())
		assert.Equal(t, n, nn)
	}

	n, err = sdktypes.ParseSymbol("#@!")
	assert.Error(t, err)
	assert.Zero(t, n)
	assert.True(t, !n.IsValid())
	assert.True(t, !sdktypes.IsValid(n))

	assert.Equal(t, "", n.Hash())

	bs, err = json.Marshal(n)
	if assert.NoError(t, err) {
		assert.Equal(t, `""`, string(bs))
	}

	assert.NotEqual(t, n, nn)
}

// If this test fails, the hash function changed. This will cause incompability with existing data.
func TestStableStringHash(t *testing.T) {
	n := kittehs.Must1(sdktypes.ParseSymbol("test"))
	assert.Equal(t, "e1b996a9b8f4e2ca87565411d377817dcecc4b7114ad4b0b6dcc42093b592c74", n.Hash())
}
