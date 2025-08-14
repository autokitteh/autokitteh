package pysdk_test

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	pysdk "go.autokitteh.dev/autokitteh/runtimes/pythonrt/py-sdk"
)

func TestDependencies(t *testing.T) {
	deps := pysdk.Dependencies()
	t.Logf("Dependencies: %v", deps)
	assert.NotEmpty(t, deps)

	for _, dep := range deps {
		assert.NotEmpty(t, dep)
		assert.Regexp(t, `^[a-zA-Z0-9_.-]+$`, dep)
	}

	// sanity checks
	assert.Less(t, sort.SearchStrings(deps, "linear"), len(deps))
	assert.Less(t, sort.SearchStrings(deps, "slack"), len(deps))
}

func TestClientNames(t *testing.T) {
	names := pysdk.ClientNames()
	t.Logf("Client names: %v", names)
	assert.NotEmpty(t, names)

	// sanity checks
	assert.Less(t, sort.SearchStrings(names, "requests"), len(names))
	assert.Less(t, sort.SearchStrings(names, "boto3"), len(names))
}
