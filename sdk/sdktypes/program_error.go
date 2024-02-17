package sdktypes

import (
	"errors"
	"fmt"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	programv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/program/v1"
)

type ProgramErrorPB = programv1.Error

type ProgramError = *object[*ProgramErrorPB]

var (
	ProgramErrorFromProto       = makeFromProto(validateProgramError)
	StrictProgramErrorFromProto = makeFromProto(strictValidateProgramError)
	ToStrictProgramError        = makeWithValidator(strictValidateProgramError)
)

func strictValidateProgramError(pb *programv1.Error) error {
	if err := ensureNotEmpty(pb.Message); err != nil {
		return err
	}

	return validateProgramError(pb)
}

func validateProgramError(pb *programv1.Error) error {
	if pb == nil {
		return nil
	}

	_, err := kittehs.ValidateList(pb.Callstack, validateCallFrame)
	return err
}

func GetProgramErrorMessage(e ProgramError) string {
	if e == nil {
		return ""
	}

	return e.pb.Message
}

func GetProgramErrorCallStack(e ProgramError) []CallFrame {
	if e == nil {
		return nil
	}

	return kittehs.Transform(e.pb.Callstack, kittehs.Must11(CallFrameFromProto))
}

// i=0: innermost. i=-1: outermost.
func GetProgramErrorCallFrameAt(e ProgramError, i int) CallFrame {
	if e == nil || len(e.pb.Callstack) == 0 {
		return nil
	}

	if i < 0 {
		i = len(e.pb.Callstack) + i
	}

	return kittehs.Must1(CallFrameFromProto(e.pb.Callstack[i]))
}

func GetProgramErrorOutermostCallFrame(e ProgramError) CallFrame {
	return GetProgramErrorCallFrameAt(e, 0)
}

func GetProgramErrorInnermostCallFrame(e ProgramError) CallFrame {
	return GetProgramErrorCallFrameAt(e, -1)
}

func GetProgramErrorExtra(e ProgramError) map[string]string {
	if e == nil {
		return nil
	}

	return e.pb.Extra
}

func NewProgramError(msg string, callstack []CallFrame, extra map[string]string) (ProgramError, error) {
	return ProgramErrorFromProto(
		&ProgramErrorPB{
			Message:   msg,
			Extra:     extra,
			Callstack: kittehs.Transform(callstack, func(f CallFrame) *CallFramePB { return f.ToProto() }),
		},
	)
}

type ProgramErrorAsError struct{ ProgramError }

func (e ProgramErrorAsError) Error() string {
	if e.ProgramError == nil {
		return ""
	}

	var b strings.Builder

	b.WriteString(e.pb.Message)

	for i, f := range e.pb.Callstack {
		b.WriteString(fmt.Sprintf("\n [%d]", i))
		if f.Location != nil {
			b.WriteString(" ")
			b.WriteString(GetCodeLocationCanonicalString(kittehs.Must1(CodeLocationFromProto(f.Location))))
		}
		if f.Name != "" {
			b.WriteString(" ")
			b.WriteString(f.Name)
		}
	}

	return b.String()
}

var _ error = ProgramErrorAsError{}

func ProgramErrorToError(e ProgramError) error {
	if e == nil {
		return nil
	}

	return &ProgramErrorAsError{e}
}

func ProgramErrorFromError(err error) ProgramError {
	if err == nil {
		return nil
	}

	var perr *ProgramErrorAsError
	if errors.As(err, &perr) {
		return perr.ProgramError
	}

	return kittehs.Must1(NewProgramError(err.Error(), nil, nil))
}
