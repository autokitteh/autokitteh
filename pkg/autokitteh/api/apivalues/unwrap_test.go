//go:build unit

package apivalues

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUnwrap(t *testing.T) {
	strptr := func(s string) *string { return &s }
	intptr := func(i int64) *int64 { return &i }
	boolptr := func(b bool) *bool { return &b }
	tptr := func(t time.Time) *time.Time { return &t }
	dptr := func(d time.Duration) *time.Duration { return &d }

	tm := time.Now().UTC()
	dr := time.Second * 42

	tests := []struct {
		n        string
		in       value
		out      interface{}
		expected interface{}
		opts     []func(*unwrapOpts)
	}{
		{
			n:        "duration",
			in:       DurationValue(dr),
			out:      dptr(0),
			expected: &dr,
		},
		{
			n:        "time",
			in:       TimeValue(tm),
			out:      tptr(time.Time{}),
			expected: &tm,
		},
		{
			n:        "str",
			in:       StringValue("meow"),
			out:      strptr(""),
			expected: strptr("meow"),
		},
		{
			n:        "int",
			in:       IntegerValue(42),
			out:      intptr(0),
			expected: intptr(42),
		},
		{
			n:        "bool",
			in:       BooleanValue(true),
			out:      boolptr(false),
			expected: boolptr(true),
		},
	}

	for _, test := range tests {
		t.Run(test.n, func(t *testing.T) {
			if !assert.NoError(t, UnwrapInto(test.out, test.in, test.opts...)) {
				return
			}

			assert.EqualValues(t, test.expected, test.out)
		})
	}
}
