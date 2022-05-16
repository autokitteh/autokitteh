package idgen

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
)

type IDGenFunc func(prefix string) string

var New = NewUUID()

func NewFakeUUID(prefix string, n uint64) string {
	return prefix + fmt.Sprintf("%x", n)
}

func NewUUID() IDGenFunc {
	return func(prefix string) string {
		return prefix + strings.Map(func(r rune) rune {
			if r == '-' {
				return -1
			}

			return r
		}, uuid.NewString())
	}
}

// mk is called only once per prefix. after that, mk ret val is reused.
func NewPerPrefix(mk func() IDGenFunc) IDGenFunc {
	fs := make(map[string]IDGenFunc)
	var lock sync.Mutex

	return func(prefix string) string {
		lock.Lock()
		defer lock.Unlock()

		f, ok := fs[prefix]
		if !ok {
			f = mk()
			fs[prefix] = f
		}

		return f(prefix)
	}
}

func NewSequential(n0 uint64) IDGenFunc {
	n := n0

	return func(prefix string) string {
		return NewFakeUUID(prefix, atomic.AddUint64(&n, 1))
	}
}

func NewSequentialPerPrefix(n0 uint64) IDGenFunc {
	return NewPerPrefix(
		func() IDGenFunc {
			return NewSequential(n0)
		},
	)
}
