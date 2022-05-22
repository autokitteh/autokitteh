package lang

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/autokitteh/autokitteh/sdk/api/apiprogram"
)

var (
	ErrLangNotRegistered          = errors.New("lang not registered")
	ErrCompilerVersionMismatch    = errors.New("compiler version mismatch")
	ErrUnsupportedCompilerVersion = errors.New("unsupported compiler version")
	ErrExtensionNotRegistered     = errors.New("extension not registered")
)

type ErrMissingDependencies []*apiprogram.Path

func (err ErrMissingDependencies) Error() string {
	deps := make([]string, len(err))
	for i, dep := range err {
		deps[i] = fmt.Sprintf("%q", dep)
	}

	return fmt.Sprintf("missing dependencies: %s", strings.Join(deps, ", "))
}

//--

type ErrCanceled struct{ CallStack []*apiprogram.CallFrame }

var _ error = &ErrCanceled{}

func (e *ErrCanceled) Error() string { return "canceled" }
func (e *ErrCanceled) Unwrap() error { return context.Canceled }
