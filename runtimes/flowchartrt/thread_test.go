package flowchartrt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThread(t *testing.T) {
	compiled, err := testBuild()
	require.NoError(t, err)

	r, err := testRun(compiled)
	require.NoError(t, err)

	th, err := r.(*run).newThread(r.Values()["n1"], nil)
	require.NoError(t, err)

	// TODO

	assert.NotNil(t, th)
}
