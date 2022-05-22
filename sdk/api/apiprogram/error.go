package apiprogram

import (
	"errors"
	"fmt"
	"strings"

	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	pbprogram "github.com/autokitteh/autokitteh/api/gen/stubs/go/program"
)

type Error struct{ pb *pbprogram.Error }

var _ error = &Error{}

func (e *Error) CallStack() []*CallFrame {
	fs := make([]*CallFrame, len(e.pb.Callstack))
	for i, pbf := range e.pb.Callstack {
		fs[i] = MustCallFrameFromProto(pbf)
	}
	return fs
}

func (e *Error) Msg() string  { return e.pb.Msg }
func (e *Error) Next() *Error { return MustErrorFromProto(e.pb.Next) }
func (e *Error) Type() string { return e.pb.Type }

func (e *Error) Error() string {
	if e == nil || e.pb == nil {
		return ""
	}

	var ls []string

	for ; e != nil; e = e.Next() {
		var l string

		if typ := e.Type(); typ != "" {
			l = fmt.Sprintf("[%s] %s", typ, e.Msg())
		} else {
			l = e.Msg()
		}

		ls = append(ls, l)
		ls = append(ls, SprintCallStack(e.CallStack()))
	}

	for i, l := range ls {
		ls[i] = strings.TrimRight(l, "\n")
	}

	return strings.Join(ls, "\n")
}

func (e *Error) PB() *pbprogram.Error {
	if e == nil || e.pb == nil {
		return nil
	}

	return proto.Clone(e.pb).(*pbprogram.Error)
}

func (e *Error) Clone() *Error {
	if e == nil || e.pb == nil {
		return nil
	}

	return &Error{pb: e.PB()}
}

// this will make the return value a real "nil" error if needed.
func (e *Error) ToError() (err error) {
	if e != nil && e.pb != nil {
		return e
	}
	return
}

func MustErrorFromProto(pb *pbprogram.Error) *Error {
	if pb == nil {
		return nil
	}

	x, err := ErrorFromProto(pb)
	if err != nil {
		panic(err)
	}
	return x
}

func GOErrorFromProto(pb *pbprogram.Error) (error, error) {
	e, err := ErrorFromProto(pb)
	if err != nil {
		return nil, err
	}

	return e.ToError(), nil
}

func ErrorFromProto(pb *pbprogram.Error) (*Error, error) {
	if pb == nil {
		return nil, nil
	}

	if err := pb.Validate(); err != nil {
		return nil, err
	}

	return (&Error{pb: pb}).Clone(), nil
}

func NewErrors(errs []*Error) *Error {
	if len(errs) == 0 {
		return nil
	}

	errs_ := make([]*Error, len(errs))
	for i, err := range errs {
		err_ := *err
		errs_[i] = &err_
	}

	for i := 0; i < len(errs_)-1; i++ {
		errs_[i].pb.Next = errs_[i+1].pb
	}

	return errs_[0]
}

func NewError(msg, typ string, callstack []*CallFrame, next *Error) (*Error, error) {
	pbcs := make([]*pbprogram.CallFrame, len(callstack))
	for i, f := range callstack {
		pbcs[i] = f.PB()
	}

	return ErrorFromProto(&pbprogram.Error{
		Msg:       msg,
		Type:      typ,
		Callstack: pbcs,
		Next:      next.PB(),
	})
}

func MustNewError(msg, typ string, callstack []*CallFrame, next *Error) *Error {
	e, err := NewError(msg, typ, callstack, next)
	if err != nil {
		panic(err)
	}
	return e
}

func ImportError(err error) *Error {
	if err == nil {
		return nil
	}

	if e := (&Error{}); errors.As(err, &e) {
		return e
	}

	return MustNewError(err.Error(), "", nil, nil)
}

func ErrorFromGRPCError(err error) *Error {
	if st, ok := status.FromError(err); ok {
		for _, deet := range st.Details() {
			if pberr, ok := deet.(*pbprogram.Error); ok {
				if perr, err := ErrorFromProto(pberr); err == nil {
					return perr
				}
			}
		}
	}

	return nil
}
