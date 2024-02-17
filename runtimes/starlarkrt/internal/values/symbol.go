package values

import (
	"go.starlark.net/starlark"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

type Symbol string

var _ starlark.Value = (*Symbol)(nil)

func (s *Symbol) String() string        { return string(*s) }
func (*Symbol) Type() string            { return "symbol" }
func (*Symbol) Freeze()                 {}
func (s *Symbol) Truth() starlark.Bool  { return true }
func (s *Symbol) Hash() (uint32, error) { return kittehs.HashString32(string(*s)), nil }

func newSymbol(sym string) *Symbol { return (*Symbol)(&sym) }
