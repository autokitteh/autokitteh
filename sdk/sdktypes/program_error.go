package sdktypes

import (
	"errors"
	"fmt"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	programv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/program/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdklogger"
)

type ProgramError struct {
	object[*ProgramErrorPB, ProgramErrorTraits]
}

func init() { registerObject[ProgramError]() }

type ProgramErrorPB = programv1.Error

type ProgramErrorTraits struct{ immutableObjectTrait }

func (ProgramErrorTraits) Validate(m *ProgramErrorPB) error {
	return errors.Join(
		objectField[Value]("value", m.Value),
		objectsSliceField[CallFrame]("callstack", m.Callstack),
	)
}

func (ProgramErrorTraits) StrictValidate(m *ProgramErrorPB) error { return nil }

var InvalidProgramError ProgramError

func (e ProgramError) Extra() map[string]string { return e.read().Extra }

func (e ProgramError) CallStack() []CallFrame {
	return kittehs.Transform(e.m.Callstack, forceFromProto[CallFrame])
}

func (e ProgramError) Value() Value {
	return forceFromProto[Value](e.read().Value)
}

func (p ProgramError) ToError() (err error) {
	if !p.IsValid() {
		return
	}

	err = programError(p)
	return
}

func (e ProgramError) ErrorString() string {
	if err := e.ToError(); err != nil {
		return err.Error()
	}
	return ""
}

func ProgramErrorFromProto(m *ProgramErrorPB) (ProgramError, error) {
	return FromProto[ProgramError](m)
}

func StrictProgramErrorFromProto(m *ProgramErrorPB) (ProgramError, error) {
	return Strict(ProgramErrorFromProto(m))
}

func NewProgramError(v Value, callstack []CallFrame, extra map[string]string) ProgramError {
	return kittehs.Must1(ProgramErrorFromProto(
		&ProgramErrorPB{
			Extra:     extra,
			Value:     v.ToProto(),
			Callstack: kittehs.Transform(callstack, func(f CallFrame) *CallFramePB { return f.ToProto() }),
		},
	))
}

func FromError(err error) (ProgramError, bool) {
	var pperr programError
	if errors.As(err, &pperr) {
		return ProgramError(pperr), true
	}
	return InvalidProgramError, false
}

func WrapError(err error) ProgramError {
	if err == nil {
		return InvalidProgramError
	}

	if perr, ok := FromError(err); ok {
		return perr
	}

	return NewProgramError(NewStringValue(err.Error()), nil, nil)
}

type programError ProgramError

var _ error = programError{}

// This allows `errors.Is(err, sdkerrors.ErrProgram)“ to work.
func (e programError) Unwrap() error { return sdkerrors.ErrProgram }

func (e programError) Error() string {
	if !e.IsValid() {
		sdklogger.DPanic("invalid")
		return ""
	}

	var b strings.Builder

	if e.m.Value != nil {
		v := forceFromProto[Value](e.m.Value)

		u, err := v.Unwrap()
		if err != nil {
			u = v
		}

		b.WriteString(fmt.Sprint(u))
	}

	for i, f := range e.m.Callstack {
		b.WriteString(fmt.Sprintf("\n [%d]", i))
		if f.Location != nil {
			b.WriteString(" ")
			b.WriteString(forceFromProto[CodeLocation](f.Location).CanonicalString())
		}
		if f.Name != "" {
			b.WriteString(" ")
			b.WriteString(f.Name)
		}
	}

	return b.String()
}
