package sdktypes

import (
	"encoding/json"
	"fmt"
	"strings"
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
	const n = len(ValidIDChars)
	for i := range ValidIDChars {
		assert.Equal(
			t,
			fmt.Sprintf("prj_000000000000000000000000%c%c", ValidIDChars[(i+1)/n], ValidIDChars[(i+1)%n]),
			NewProjectID().String(),
		)
	}
}

func TestFromID(t *testing.T) {
	tid := NewTriggerID()
	id := NewEventDestinationID(tid)
	assert.Equal(t, FromID[TriggerID](id), tid)
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

// If this test fails, the hash function changed. This will cause incompatibility with existing data.
func TestStableIDHash(t *testing.T) {
	id := kittehs.Must1(ParseProjectID(pidStr))
	assert.Equal(t, testHash, id.Hash())
}

func TestNewNamedIDString(t *testing.T) {
	tests := []struct {
		in, out string
	}{
		{out: "tst_3kth0000008bcc2c7e9f98f09a"},
		{in: "meow", out: "tst_3kth00me0wef7dc93aa8e6f016"},
		{in: "meowwoofoink", out: "tst_3kthme0ww0730067fc95b3ff1e"},
		{in: "meowwoofsqueek", out: "tst_3kthme0ww0c9b069d5c3269215"},
		{in: "123@#$", out: "tst_3kth12300s632f857c8d36a0d3"},
	}

	for _, test := range tests {
		t.Run(test.in, func(t *testing.T) {
			assert.Equal(t, test.out, newNamedIDString(test.in, "tst"))
		})
	}
}

func TestSmartParseID(t *testing.T) {
	// 1. Full valid project ID should parse directly
	full := pidStr
	pid, err := SmartParseID[ProjectID](full)
	if assert.NoError(t, err) {
		assert.Equal(t, full, pid.String())
	}

	// 2. UUID form: take underlying UUID of an ID and feed just the UUID string
	orig := NewProjectID()
	u := orig.UUIDValue().String()
	pid2, err := SmartParseID[ProjectID](u)
	assert.NoError(t, err)
	assert.Equal(t, orig.UUIDValue().String(), pid2.UUIDValue().String())

	// 3. Short suffix form (drop prefix + underscore) should be accepted
	short := strings.TrimPrefix(full, "prj_")
	pid3, err := SmartParseID[ProjectID](short)
	if assert.NoError(t, err) {
		assert.Equal(t, full, pid3.String())
	}

	// 4. Invalid formats
	_, err = SmartParseID[ProjectID]("not-a-valid-id!!!")
	assert.Error(t, err)

	id, err := SmartParseID[ProjectID]("")
	assert.NoError(t, err)
	assert.False(t, id.IsValid())
}
