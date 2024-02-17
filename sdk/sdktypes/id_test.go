package sdktypes

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

var (
	projectID = kittehs.Must1(ParseProjectID("p:00000000000000000000000000000001"))
	envID     = kittehs.Must1(ParseEnvID("e:00000000000000000000000000000001"))
)

func TestParseAnyID(t *testing.T) {
	id, err := ParseAnyID("")

	if assert.NoError(t, err) {
		assert.Nil(t, id)
	}

	testParseAnyID(t, ParseAnyID)
}

func TestStrictParseAnyID(t *testing.T) {
	id, err := StrictParseAnyID("")

	if assert.Error(t, err) {
		assert.True(t, errors.Is(err, sdkerrors.ErrInvalidArgument))
	}

	assert.Nil(t, id)

	testParseAnyID(t, StrictParseAnyID)
}

func testParseAnyID(t *testing.T, parse func(string) (ID, error)) {
	tests := []struct {
		name string
		in   string // if out != nil && empty, set to out.String().
		out  ID
		err  bool // if true, ErrInvalidArgument is expected.
	}{
		{
			name: "project-id",
			out:  projectID,
		},
		{
			name: "env-id",
			out:  envID,
		},
		{
			name: "unknown",
			in:   "x:00000000000000000000000000000001",
			err:  true,
		},
		{
			name: "length",
			in:   "u:1",
			err:  true,
		},
		{
			name: "nonhex",
			in:   "u:0000000000000000000000000000meow",
			err:  true,
		},
		{
			name: "nokind",
			in:   ":00000000000000000000000000000001",
			err:  true,
		},
		{
			name: "nouuid",
			in:   "u:",
			err:  true,
		},
		{
			name: "wat",
			in:   "meow",
			err:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			in := test.in

			if test.out != nil && in == "" {
				in = test.out.String()
			}

			id, err := parse(in)

			if test.err {
				assert.Nil(t, id)
				if assert.Error(t, err) {
					assert.True(t, errors.Is(err, sdkerrors.ErrInvalidArgument))
				}

				return
			}

			assert.NoError(t, err)

			if test.out != nil {
				assert.NotNil(t, id)
				assert.Equal(t, test.out.String(), id.String())
			} else {
				assert.Nil(t, id)
			}
		})
	}
}

func TestIsID(t *testing.T) {
	tests := []struct {
		name, in string
		out      bool
		valid    bool
	}{
		{
			name: "empty",
		},
		{
			name:  "pid",
			in:    projectID.String(),
			out:   true,
			valid: true,
		},
		{
			name:  "eid",
			in:    envID.String(),
			out:   true,
			valid: true,
		},
		{
			name: "nodelim",
			in:   "meow",
		},
		{
			name: "delim",
			in:   ":",
			out:  true,
		},
		{
			name: "delimish",
			in:   "x:y",
			out:  true,
		},
		{
			name:  "unknown",
			in:    "meow:00000000000000000000000000000001",
			out:   true,
			valid: true,
		},
		{
			name:  "wrong-len",
			in:    "meow:1",
			out:   true,
			valid: true,
		},
		{
			name: "nothex",
			in:   "u:0000000000000000000000000000meow",
			out:  true,
		},
		{
			name: "nokind",
			in:   ":00000000000000000000000000001234",
			out:  true,
		},
		{
			name: "nouuid",
			in:   "u:",
			out:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, IsID(test.in), test.out)
			assert.Equal(t, IsValidID(test.in), test.valid)
		})
	}
}

func TestSplitRawID(t *testing.T) {
	tests := []struct {
		in         string
		kind, data string
	}{
		{},
		{
			in:   "u",
			kind: "u",
		},
		{
			in:   "u:",
			kind: "u",
		},
		{
			in:   "u:0123",
			kind: "u",
			data: "0123",
		},
		{
			in:   ":0123",
			data: "0123",
		},
	}

	for _, test := range tests {
		t.Run(test.in, func(t *testing.T) {
			kind, data, ok := SplitRawID(test.in)
			assert.Equal(t, test.kind, kind)
			assert.Equal(t, test.data, data)
			assert.Equal(t, IsID(test.in), ok)
		})
	}
}
