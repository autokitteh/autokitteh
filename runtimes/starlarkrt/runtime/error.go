package runtime

import (
	"errors"
	"fmt"

	"go.starlark.net/resolve"
	"go.starlark.net/starlark"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func translateError(err error, extra map[string]string) error {
	if err == nil {
		return nil
	}

	convErr := func(cerr, err error) error {
		return fmt.Errorf("[cannot convert to program error: %v] %w", cerr, err)
	}

	if resolveErrorList := (resolve.ErrorList{}); errors.As(err, &resolveErrorList) {
		return translateError(resolveErrorList[0], map[string]string{
			"all": resolveErrorList.Error(),
		})
	}

	if resolveError := (resolve.Error{}); errors.As(err, &resolveError) {
		f, cerr := sdktypes.CallFrameFromProto(&sdktypes.CallFramePB{
			Location: &sdktypes.CodeLocationPB{
				Path: resolveError.Pos.Filename(),
				Row:  uint32(resolveError.Pos.Line),
				Col:  uint32(resolveError.Pos.Col),
			},
		})
		if cerr != nil {
			return convErr(cerr, err)
		}

		extra, _ := kittehs.JoinMaps(
			map[string]string{
				"raw":  resolveError.Error(),
				"type": "resolve",
			},
			extra,
		)

		perr := sdktypes.NewProgramError(sdktypes.NewStringValue(resolveError.Msg), []sdktypes.CallFrame{f}, extra)
		return perr.ToError()
	} else if evalErr := (&starlark.EvalError{}); errors.As(err, &evalErr) {
		callstack, cerr := kittehs.TransformError(
			evalErr.CallStack,
			func(f starlark.CallFrame) (sdktypes.CallFrame, error) {
				return sdktypes.CallFrameFromProto(&sdktypes.CallFramePB{
					Name: f.Name,
					Location: &sdktypes.CodeLocationPB{
						Path: f.Pos.Filename(),
						Row:  uint32(f.Pos.Line),
						Col:  uint32(f.Pos.Col),
						Name: f.Name,
					},
				})
			},
		)
		if cerr != nil {
			return convErr(cerr, err)
		}

		extra, _ := kittehs.JoinMaps(
			map[string]string{
				"raw":   evalErr.Error(),
				"type":  "eval",
				"cause": fmt.Sprintf("%v", evalErr.Unwrap()),
			},
			extra,
		)

		perr := sdktypes.NewProgramError(sdktypes.NewStringValue(evalErr.Msg), callstack, extra)
		return perr.ToError()
	}

	return err
}
