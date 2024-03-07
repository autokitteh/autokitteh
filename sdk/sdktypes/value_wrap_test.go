package sdktypes_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	w = sdktypes.DefaultValueWrapper

	iv           = sdktypes.NewIntegerValue(42)
	nothing      = sdktypes.Nothing
	intsList     = sdktypes.NewListValue([]sdktypes.Value{sdktypes.NewIntegerValue(1), sdktypes.NewIntegerValue(2), sdktypes.NewIntegerValue(3)})
	intsSet      = sdktypes.NewSetValue([]sdktypes.Value{sdktypes.NewIntegerValue(1), sdktypes.NewIntegerValue(2), sdktypes.NewIntegerValue(3)})
	mixedList    = sdktypes.NewListValue([]sdktypes.Value{sdktypes.NewIntegerValue(1), sdktypes.NewStringValue("meow"), sdktypes.NewFloatValue(1.2)})
	stringIntMap = map[string]sdktypes.Value{
		"one":   sdktypes.NewIntegerValue(1),
		"two":   sdktypes.NewIntegerValue(2),
		"three": sdktypes.NewIntegerValue(3),
	}
	stringIntDict = sdktypes.NewDictValueFromStringMap(stringIntMap)
	intsListList  = sdktypes.NewListValue([]sdktypes.Value{
		sdktypes.NewListValue([]sdktypes.Value{sdktypes.NewIntegerValue(1)}),
		sdktypes.NewListValue([]sdktypes.Value{sdktypes.NewIntegerValue(1), sdktypes.NewIntegerValue(2)}),
		sdktypes.NewListValue([]sdktypes.Value{sdktypes.NewIntegerValue(0), sdktypes.NewIntegerValue(1), sdktypes.NewIntegerValue(2)}),
	})

	stringIntStruct = kittehs.Must1(sdktypes.NewStructValue(sdktypes.NewStringValue("ctor"), stringIntMap))
)

// TODO: These tests are not exhaustive.
func TestValueWrapper(t *testing.T) {
	type Wstr string
	type Wint int
	type Wfloat64 float64

	tests := []struct {
		in  any
		w   sdktypes.Value
		unw any
	}{
		{
			in:  42,
			w:   sdktypes.NewIntegerValue(42),
			unw: int64(42),
		},
		{
			in:  Wint(42),
			w:   sdktypes.NewIntegerValue(42),
			unw: int64(42),
		},
		{
			in:  "meow",
			w:   sdktypes.NewStringValue("meow"),
			unw: "meow",
		},
		{
			in:  Wstr("meow"),
			w:   sdktypes.NewStringValue("meow"),
			unw: "meow",
		},
		{
			in:  struct{}{},
			w:   sdktypes.Nothing,
			unw: struct{}{},
		},
		{
			in:  9.0,
			w:   sdktypes.NewFloatValue(9.0),
			unw: 9.0,
		},
		{
			in:  9.1,
			w:   sdktypes.NewFloatValue(9.1),
			unw: 9.1,
		},
		{
			in:  Wfloat64(42.1),
			w:   sdktypes.NewFloatValue(42.1),
			unw: 42.1,
		},
		{
			in:  Wfloat64(42.0),
			w:   sdktypes.NewIntegerValue(42),
			unw: int64(42),
		},
		{
			in:  []byte{1, 2, 3},
			w:   sdktypes.NewBytesValue([]byte{1, 2, 3}),
			unw: []byte{1, 2, 3},
		},
		{
			in: []int{1, 2, 3},
			w: sdktypes.NewListValue([]sdktypes.Value{
				sdktypes.NewIntegerValue(1),
				sdktypes.NewIntegerValue(2),
				sdktypes.NewIntegerValue(3),
			}),
			unw: []any{int64(1), int64(2), int64(3)},
		},
		{
			in: map[string]int{"meow": 1, "woof": 2},
			w: kittehs.Must1(sdktypes.NewDictValue([]sdktypes.DictItem{
				{
					K: sdktypes.NewStringValue("meow"),
					V: sdktypes.NewIntegerValue(1),
				},
				{
					K: sdktypes.NewStringValue("woof"),
					V: sdktypes.NewIntegerValue(2),
				},
			})),
			unw: map[any]any{"meow": int64(1), "woof": int64(2)},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v", test.in), func(t *testing.T) {
			var w sdktypes.ValueWrapper
			v, err := w.Wrap(test.in)
			if !assert.NoError(t, err) {
				return
			}

			assert.Equal(t, test.w.ToProto(), v.ToProto())

			unw, err := w.Unwrap(v)
			if assert.NoError(t, err) {
				assert.Equal(t, test.unw, unw)
			}
		})
	}
}

func TestUnwrapIntoScalars(t *testing.T) {
	var i int
	if assert.NoError(t, w.UnwrapInto(&i, iv)) {
		assert.Equal(t, 42, i)
	}

	var i64 int
	if assert.NoError(t, w.UnwrapInto(&i64, iv)) {
		assert.Equal(t, 42, i64)
	}

	var s string
	if assert.NoError(t, w.UnwrapInto(&s, iv)) {
		// yeah yeah don't blame me, blame reflect for doing this.
		assert.Equal(t, "*" /* ASCII 42 */, s)
	}

	if assert.NoError(t, w.UnwrapInto(&s, sdktypes.NewSymbolValue(kittehs.Must1(sdktypes.ParseSymbol("meow"))))) {
		assert.Equal(t, "meow", s)
	}
}

func TestUnwrapIntoPtrs(t *testing.T) {
	one := 1
	var pint *int = &one
	if assert.NoError(t, w.UnwrapInto(&pint, nothing)) {
		assert.Nil(t, pint)
	}

	pint = nil
	if assert.NoError(t, w.UnwrapInto(&pint, iv)) && assert.NotNil(t, pint) {
		assert.Equal(t, 42, *pint)
	}

	pint = new(int)
	if assert.NoError(t, w.UnwrapInto(&pint, iv)) && assert.NotNil(t, pint) {
		assert.Equal(t, 42, *pint)
	}
}

func TestUnwrapIntoCollections(t *testing.T) {
	var bs []byte
	assert.EqualError(t, w.UnwrapInto(&bs, iv), "cannot unwrap into []uint8")

	var is []int
	if assert.NoError(t, w.UnwrapInto(&is, intsList)) {
		assert.Equal(t, []int{1, 2, 3}, is)
	}

	var iis [][]int
	if assert.NoError(t, w.UnwrapInto(&iis, intsListList)) {
		assert.Equal(t, [][]int{{1}, {1, 2}, {0, 1, 2}}, iis)
	}

	assert.EqualError(t, w.UnwrapInto(&is, mixedList), "1: cannot unwrap into int")

	var stl []struct{}
	assert.EqualError(t, w.UnwrapInto(&stl, intsList), "0: cannot unwrap into struct {}")

	var s string
	assert.EqualError(t, w.UnwrapInto(&s, intsList), "cannot unwrap into string")

	var arr3 [3]int
	if assert.NoError(t, w.UnwrapInto(&arr3, intsList)) {
		assert.Equal(t, [3]int{1, 2, 3}, arr3)
	}

	var arr2 [2]int
	assert.EqualError(t, w.UnwrapInto(&arr2, intsList), "cannot unwrap into [2]int")

	var m map[string]int
	if assert.NoError(t, w.UnwrapInto(&m, stringIntDict)) {
		assert.Equal(t, map[string]int{"one": 1, "two": 2, "three": 3}, m)
	}

	var boolm map[int]bool
	if assert.NoError(t, w.UnwrapInto(&boolm, intsSet)) {
		assert.Equal(t, map[int]bool{1: true, 2: true, 3: true}, boolm)
	}

	assert.Error(t, w.UnwrapInto(&boolm, intsList))

	clear(m)

	if assert.NoError(t, w.UnwrapInto(&m, stringIntStruct)) {
		assert.Equal(t, map[string]int{"one": 1, "two": 2, "three": 3}, m)
	}

	var v sdktypes.Value
	if assert.NoError(t, w.UnwrapInto(&v, iv)) {
		assert.Equal(t, int64(42), v.GetInteger().Value())
	}

	var a any
	if assert.NoError(t, w.UnwrapInto(&a, iv)) {
		assert.Equal(t, int64(42), a)
	}
}

func TestUnwrapIntoStructs(t *testing.T) {
	var st struct{}
	assert.EqualError(t, w.UnwrapInto(&st, iv), "cannot unwrap into struct {}")

	var st1, zerost struct {
		One, Two, Three int
	}
	if assert.NoError(t, w.UnwrapInto(&st1, stringIntDict)) {
		assert.Equal(t, 1, st1.One)
		assert.Equal(t, 2, st1.Two)
		assert.Equal(t, 3, st1.Three)
	}

	st1 = zerost

	if assert.NoError(t, w.UnwrapInto(&st1, stringIntStruct)) {
		assert.Equal(t, 1, st1.One)
		assert.Equal(t, 2, st1.Two)
		assert.Equal(t, 3, st1.Three)
	}
}

func TestUnwrapIntoSpecials(t *testing.T) {
	var d time.Duration
	if assert.NoError(t, w.UnwrapInto(&d, sdktypes.NewStringValue("1s"))) {
		assert.Equal(t, time.Second, d)
	}

	if assert.NoError(t, w.UnwrapInto(&d, sdktypes.NewIntegerValue(3))) {
		assert.Equal(t, time.Second*3, d)
	}

	var dptr *time.Duration
	if assert.NoError(t, w.UnwrapInto(&dptr, sdktypes.NewStringValue("1s"))) && assert.NotNil(t, dptr) {
		assert.Equal(t, time.Second, *dptr)
	}

	var tm time.Time
	if assert.NoError(t, w.UnwrapInto(&tm, sdktypes.NewStringValue("1/1/23 18:32"))) {
		assert.Equal(t, time.Date(2023, time.January, 1, 18, 32, 0, 0, time.UTC), tm)
	}
}

func TestUnwrapIntoValue(t *testing.T) {
	var i sdktypes.IntegerValue
	if assert.NoError(t, w.UnwrapInto(&i, iv)) {
		assert.Equal(t, i.String(), "v:42")
	}

	var d sdktypes.DictValue
	if assert.NoError(t, w.UnwrapInto(&d, stringIntDict)) {
		var m map[string]int
		if assert.NoError(t, w.UnwrapInto(&m, sdktypes.NewValue(d))) {
			assert.Equal(t, map[string]int{"one": 1, "two": 2, "three": 3}, m)
		}
	}
}

func TestUnwrapIntoKitchenSink(t *testing.T) {
	type Y struct {
		Z string
	}

	type X struct {
		I64       int64
		S         string
		B         bool
		F         float64
		A2        [2]string
		M         map[int]string
		Set       map[string]bool
		Sl        []float32
		StsA      [3]Y
		StsS      []Y
		Bptr      *bool
		Sptr      *Y
		SptrS     []*Y
		UnsetI    int
		UnsetIPtr *int
		D         time.Duration
		InS       struct {
			T time.Time
		}
	}

	True := true

	in := X{
		I64:   42,
		S:     "meow",
		B:     true,
		F:     4.2,
		A2:    [2]string{"meow", "woof"},
		M:     map[int]string{1: "one", 7: "seven"},
		Set:   map[string]bool{"one": true, "two": false},
		Sl:    []float32{1.2, 3.4},
		StsA:  [3]Y{{Z: "first"}, {Z: "second"}, {Z: "third"}},
		StsS:  []Y{{Z: "uno"}, {Z: "dos"}, {Z: "tres"}},
		Bptr:  &True,
		Sptr:  &Y{Z: "neo"},
		SptrS: []*Y{{Z: "meow"}, nil, {Z: "woof"}},
		D:     time.Hour,
		InS:   struct{ T time.Time }{T: time.Date(2023, time.January, 1, 18, 32, 0, 0, time.UTC)},
	}

	w := sdktypes.DefaultValueWrapper

	var x X

	wx := kittehs.Must1(w.Wrap(in))

	if assert.NoError(t, w.UnwrapInto(&x, wx)) {
		assert.Equal(t, in, x)
	}
}
