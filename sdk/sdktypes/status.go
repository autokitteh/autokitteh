package sdktypes

import (
	"errors"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	commonv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/common/v1"
)

type Status struct {
	object[*StatusPB, StatusTraits]
}

func init() { registerObject[Status]() }

var InvalidStatus Status

type StatusPB = commonv1.Status

type StatusTraits struct{ immutableObjectTrait }

func (StatusTraits) Validate(m *StatusPB) error {
	return errors.Join(
		enumField[StatusCode]("code", m.Code),
	)
}

func (StatusTraits) StrictValidate(m *StatusPB) error { return nil }

func StatusFromProto(m *StatusPB) (Status, error) { return FromProto[Status](m) }
func StrictStatusFromProto(m *StatusPB) (Status, error) {
	return Strict(StatusFromProto(m))
}

func (p Status) Code() StatusCode {
	return forceEnumFromProto[StatusCode](p.read().Code)
}

func (p Status) WithState(s StatusCode) Status {
	return Status{p.forceUpdate(func(pb *StatusPB) { pb.Code = s.ToProto() })}
}

func (p Status) Message() string { return p.read().Message }

func NewStatus(code StatusCode, msg string) Status {
	return kittehs.Must1(StatusFromProto(&StatusPB{
		Code:    code.ToProto(),
		Message: msg,
	}))
}
