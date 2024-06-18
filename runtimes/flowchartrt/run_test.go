package flowchartrt

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func testRun(compiled map[string][]byte) (sdkservices.Run, error) {
	return rt{}.Run(context.Background(), sdktypes.NewRunID(), "main.flowchart.yaml", compiled, nil, nil)
}

func TestRun(t *testing.T) {
	compiled, err := testBuild()
	require.NoError(t, err)

	r, err := testRun(compiled)
	require.NoError(t, err)

	// TODO
	require.NotNil(t, r)

	assert.NotNil(t, r.Values()["n1"])
	assert.NotNil(t, r.Values()["n2"])
}
