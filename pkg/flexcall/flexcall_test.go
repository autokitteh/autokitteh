//go:build unit

package flexcall

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCall(t *testing.T) {
	tests := []struct {
		n       string
		f       interface{}
		ins     []interface{}
		outs    []interface{}
		optOuts []interface{}
		err     error
		optErr  error
	}{
		{
			n: "empty",
			f: func() {},
		},
		{
			n:   "single in",
			f:   func(int) {},
			ins: []interface{}{int(1)},
		},
		{
			n:    "single in, error out",
			f:    func(int) error { return nil },
			ins:  []interface{}{int(1)},
			outs: []interface{}{error(nil)},
		},
		{
			n:       "not exists, int",
			f:       func(x int) (int, error) { return x, nil },
			err:     ErrUnmatched,
			optOuts: []interface{}{int(0), error(nil)},
		},
		{
			n:       "not exists, ptr",
			f:       func(x *int) (*int, error) { return x, nil },
			err:     ErrUnmatched,
			optOuts: []interface{}{(*int)(nil), error(nil)},
		},
		{
			n:       "not exists, interface",
			f:       func(c context.Context) (context.Context, error) { return c, nil },
			err:     ErrUnmatched,
			optOuts: []interface{}{(context.Context)(nil), error(nil)},
		},
		{
			n:   "multi in",
			f:   func(int, string, float64) {},
			ins: []interface{}{int(1), "one", 1.0},
		},
		{
			n:    "mul",
			f:    func(a int, b float64) float64 { return float64(a) * b },
			ins:  []interface{}{int(2), 4.2},
			outs: []interface{}{8.4},
		},
		{
			n:    "multi out",
			f:    func() (int, int) { return 1, 2 },
			outs: []interface{}{int(1), int(2)},
		},
		{
			n:    "multi in/out",
			f:    func(a string, b int) (int, string) { return b + 1, a + "1" },
			ins:  []interface{}{int(1), "one"},
			outs: []interface{}{int(2), "one1"},
		},
		{
			n:   "context (interface)",
			f:   func(context.Context) {},
			ins: []interface{}{context.Background()},
		},
	}

	for _, test := range tests {
		t.Run(test.n, func(t *testing.T) {
			outs, err := Call(test.f, test.ins...)

			if test.err != nil {
				assert.Nil(t, outs)
				assert.True(t, errors.Is(err, test.err), err)
			} else {
				assert.NoError(t, err)
				assert.EqualValues(t, test.outs, outs)
			}

			if test.optOuts == nil {
				test.optOuts = test.outs
			}

			outs, err = CallOptional(test.f, test.ins...)

			t.Log(test.optErr)

			if test.optErr != nil {
				assert.Nil(t, outs)
				assert.True(t, errors.Is(err, test.optErr))
			} else {
				assert.NoError(t, err)
				assert.EqualValues(t, test.optOuts, outs)
			}
		})
	}
}

func TestExtractError(t *testing.T) {
	err1, err2 := errors.New("err1"), errors.New("err2")

	var nilErr error

	tests := []struct {
		n    string
		f    interface{}
		ins  []interface{}
		err  error
		outs []interface{}
	}{
		{
			n: "empty",
			f: func() {},
		},
		{
			n:    "just error",
			f:    func() error { return nil },
			ins:  []interface{}{err1},
			err:  err1,
			outs: []interface{}{},
		},
		{
			n:    "just nil error",
			f:    func() error { return nil },
			ins:  []interface{}{nilErr},
			err:  nilErr,
			outs: []interface{}{},
		},
		{
			n:    "multi with error",
			f:    func() (int, string, error) { return 1, "", nil },
			ins:  []interface{}{int(1), "", err1},
			err:  err1,
			outs: []interface{}{int(1), ""},
		},
		{
			n:    "multi without error",
			f:    func() (int, string) { return 1, "" },
			ins:  []interface{}{int(1), ""},
			outs: []interface{}{int(1), ""},
		},
		{
			n:    "error not last",
			f:    func() (error, string) { return nil, "" },
			ins:  []interface{}{err1, ""},
			outs: []interface{}{err1, ""},
		},
		{
			n:    "multi errs",
			f:    func() (error, error) { return nil, nil },
			ins:  []interface{}{err1, err2},
			err:  err2,
			outs: []interface{}{err1},
		},
		{
			n:    "multi errs with ins",
			f:    func() (string, error, error) { return "", nil, nil },
			ins:  []interface{}{"", err1, err2},
			err:  err2,
			outs: []interface{}{"", err1},
		},
	}

	for _, test := range tests {
		t.Run(test.n, func(t *testing.T) {
			outs, err := ExtractError(test.f, test.ins)

			assert.Equal(t, test.err, err)
			assert.EqualValues(t, test.outs, outs)
		})
	}
}
