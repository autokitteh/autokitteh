//go:build unit

package apivalues

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type sortByIntKey struct{ vs []*DictItem }

func (s sortByIntKey) Len() int      { return len(s.vs) }
func (s sortByIntKey) Swap(i, j int) { s.vs[i], s.vs[j] = s.vs[j], s.vs[i] }
func (s sortByIntKey) Less(i, j int) bool {
	return s.vs[i].K.Get().(IntegerValue) < s.vs[j].K.Get().(IntegerValue)
}

func TestWrap(t *testing.T) {
	V := MustNewValue

	true_, one := true, 1
	tm := time.Now().UTC()
	dr := time.Second * 42

	tests := []struct {
		n    string
		in   interface{}
		out  value
		opts []func(*wrapOpts)
	}{
		{
			n:   "duration",
			in:  dr,
			out: DurationValue(dr),
		},
		{
			n:   "duration ptr",
			in:  &dr,
			out: DurationValue(dr),
		},
		{
			n:   "time",
			in:  tm,
			out: TimeValue(tm),
		},
		{
			n:   "time ptr",
			in:  &tm,
			out: TimeValue(tm),
		},
		{
			n:   "bool",
			in:  true,
			out: BooleanValue(true),
		},
		{
			n:   "bool ptr",
			in:  &true_,
			out: BooleanValue(true),
		},
		{
			n:   "int",
			in:  1,
			out: IntegerValue(1),
		},
		{
			n:   "int ptr",
			in:  &one,
			out: IntegerValue(1),
		},
		{
			n:   "slice",
			in:  []int{1, 2},
			out: ListValue([]*Value{V(IntegerValue(1)), V(IntegerValue(2))}),
		},
		{
			n:   "array",
			in:  [2]int{1, 2},
			out: ListValue([]*Value{V(IntegerValue(1)), V(IntegerValue(2))}),
		},
		{
			n:  "slice of slices",
			in: [][]string{{"meow", "woof"}, {"oink", "moo", "chirp"}, {}},
			out: ListValue(
				[]*Value{
					V(ListValue([]*Value{V(StringValue("meow")), V(StringValue("woof"))})),
					V(ListValue([]*Value{V(StringValue("oink")), V(StringValue("moo")), V(StringValue("chirp"))})),
					V(ListValue([]*Value{})),
				},
			),
		},
		{
			n:  "map",
			in: map[int]*int{42: &one, 3: nil},
			out: DictValue(
				[]*DictItem{
					{K: V(IntegerValue(3)), V: None},
					{K: V(IntegerValue(42)), V: V(IntegerValue(1))},
				},
			),
			opts: []func(*wrapOpts){
				WithDictSorter(func(vs []*DictItem) {
					sort.Sort(sortByIntKey{vs: vs})
				}),
			},
		},
		{
			n:   "nil",
			in:  (*int)(nil),
			out: NoneValue{},
		},
		{
			n:   "empty struct",
			in:  (*struct{})(nil),
			out: NoneValue{},
		},
	}

	for _, test := range tests {
		t.Run(test.n, func(t *testing.T) {
			v, err := Wrap(test.in, test.opts...)
			if !assert.NoError(t, err) {
				return
			}

			if !assert.EqualValues(t, test.out, v.Get()) {
				fmt.Printf("%v %v", test.out, v.Get())
			}
		})
	}
}
