package sdktypes_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var _ sdktypes.ID = sdktypes.InvalidProjectID

const (
	pidStr = "prj_01hqmf3sesfg0ayxnfqyraxn7n"
	hash   = "f562c1923f5bc759f218236917107e3abbc5da65515d459c4fa91d624c3e739e"
)

func TestSequentialID(t *testing.T) {
	sdktypes.SetIDGenerator(sdktypes.NewSequentialIDGeneratorForTesting(0))
	const chars = "0123456789abcdefghjkmnpqrstvwxyz"
	const n = len(chars)
	for i := range chars {
		assert.Equal(
			t,
			fmt.Sprintf("prj_000000000000000000000000%c%c", chars[(i+1)/n], chars[(i+1)%n]),
			sdktypes.NewProjectID().String(),
		)
	}
}

func TestID(t *testing.T) {
	zero, err := sdktypes.ParseProjectID("")
	if assert.NoError(t, err) {
		assert.False(t, zero.IsValid())
		assert.False(t, sdktypes.IsValid(zero))
	}

	assert.Error(t, zero.Strict())

	z, err := sdktypes.Strict(sdktypes.ParseProjectID(""))
	assert.Error(t, err)
	assert.False(t, z.IsValid())

	pid, err := sdktypes.ParseProjectID("prj_01hqmf3sesfg0ayxnfqyraxn7n")
	if assert.NoError(t, err) {
		assert.Equal(t, hash, pid.Hash())
	}

	bs, err := json.Marshal(pid)
	if assert.NoError(t, err) {
		assert.Equal(t, `"`+pidStr+`"`, string(bs))
	}

	assert.NotEqual(t, zero, pid)

	var pid2 sdktypes.ProjectID
	if assert.NoError(t, pid2.UnmarshalJSON([]byte(`"`+pidStr+`"`))) {
		assert.Equal(t, pid, pid2)
	}

	assert.NotEqual(t, zero, pid2)
}

// If this test fails, the hash function changed. This will cause incompability with existing data.
func TestStableIDHash(t *testing.T) {
	id := kittehs.Must1(sdktypes.ParseProjectID(pidStr))
	assert.Equal(t, hash, id.Hash())
}
