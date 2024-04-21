package sdktypes

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

var _ ID = InvalidProjectID

const (
	pidStr   = "prj_01hqmf3sesfg0ayxnfqyraxn7n"
	testHash = "f562c1923f5bc759f218236917107e3abbc5da65515d459c4fa91d624c3e739e"
)

func TestSequentialID(t *testing.T) {
	SetIDGenerator(NewSequentialIDGeneratorForTesting(0))
	const n = len(validIDChars)
	for i := range validIDChars {
		assert.Equal(
			t,
			fmt.Sprintf("prj_000000000000000000000000%c%c", validIDChars[(i+1)/n], validIDChars[(i+1)%n]),
			NewProjectID().String(),
		)
	}
}

func TestID(t *testing.T) {
	zero, err := ParseProjectID("")
	if assert.NoError(t, err) {
		assert.False(t, zero.IsValid())
		assert.False(t, IsValid(zero))
	}

	assert.Error(t, zero.Strict())

	z, err := Strict(ParseProjectID(""))
	assert.Error(t, err)
	assert.False(t, z.IsValid())

	pid, err := ParseProjectID("prj_01hqmf3sesfg0ayxnfqyraxn7n")
	if assert.NoError(t, err) {
		assert.Equal(t, testHash, pid.Hash())
	}

	bs, err := json.Marshal(pid)
	if assert.NoError(t, err) {
		assert.Equal(t, `"`+pidStr+`"`, string(bs))
	}

	assert.NotEqual(t, zero, pid)

	var pid2 ProjectID
	if assert.NoError(t, pid2.UnmarshalJSON([]byte(`"`+pidStr+`"`))) {
		assert.Equal(t, pid, pid2)
	}

	assert.NotEqual(t, zero, pid2)
}

// If this test fails, the hash function changed. This will cause incompability with existing data.
func TestStableIDHash(t *testing.T) {
	id := kittehs.Must1(ParseProjectID(pidStr))
	assert.Equal(t, testHash, id.Hash())
}
