package langstarlark

import (
	"context"
	"errors"
	"fmt"

	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"

	"go.autokitteh.dev/sdk/api/apiprogram"
	"github.com/autokitteh/autokitteh/internal/pkg/lang"
)

func newCallFrame(name string, pos *syntax.Position) *apiprogram.CallFrame {
	return apiprogram.MustNewCallFrame(
		name,
		apiprogram.MustNewLocation(
			apiprogram.ParsePathStringOr(pos.Filename(), apiprogram.MustParsePathString("unknown")),
			pos.Line,
			pos.Col,
		),
	)
}

func resolveErrToError(err *resolve.Error) *apiprogram.Error {
	return apiprogram.MustNewError(
		err.Msg,
		"resolve",
		[]*apiprogram.CallFrame{newCallFrame("", &err.Pos)},
		nil,
	)
}

func errf(f string, vs ...interface{}) error { return translateError(fmt.Errorf(f, vs...)) }

func translateError(err error) error {
	if err == nil {
		return nil
	}

	if slErr := (&starlark.EvalError{}); errors.As(err, &slErr) {
		fs := make([]*apiprogram.CallFrame, len(slErr.CallStack))
		for i, f := range slErr.CallStack {
			fs[i] = newCallFrame(f.Name, &f.Pos)
		}

		if err := slErr.Unwrap(); errors.Is(err, context.Canceled) {
			return &lang.ErrCanceled{CallStack: fs}
		}

		return apiprogram.MustNewError(slErr.Msg, "eval", fs, nil)
	}

	if slErr := (syntax.Error{}); errors.As(err, &slErr) {
		return apiprogram.MustNewError(
			slErr.Msg,
			"syntax",
			[]*apiprogram.CallFrame{newCallFrame("", &slErr.Pos)},
			nil,
		)
	}

	if slErr := (&resolve.Error{}); errors.As(err, &slErr) {
		return resolveErrToError(slErr)
	}

	if errList := (resolve.ErrorList{}); errors.As(err, &errList) {
		errs := make([]*apiprogram.Error, len(errList))
		for i, err := range errList {
			errs[i] = resolveErrToError(&err)
		}

		return apiprogram.NewErrors(errs)
	}

	return apiprogram.MustNewError(
		err.Error(),
		"",
		nil,
		nil,
	)
}
