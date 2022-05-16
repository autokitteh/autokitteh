package apiprogram

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"

	pbprogram "github.com/autokitteh/autokitteh/gen/proto/stubs/go/program"
)

type CallFrame struct{ pb *pbprogram.CallFrame }

func (f *CallFrame) PB() *pbprogram.CallFrame {
	if f == nil || f.pb == nil {
		return nil
	}

	return proto.Clone(f.pb).(*pbprogram.CallFrame)
}

func (f *CallFrame) Clone() *CallFrame {
	if f == nil || f.pb == nil {
		return nil
	}

	return &CallFrame{pb: f.PB()}
}

func (f *CallFrame) Name() string { return f.pb.Name }

func (f *CallFrame) Location() *Location { return MustLocationFromProto(f.pb.Loc) }

func (f *CallFrame) String() string {
	return fmt.Sprintf("%v %s", f.Location(), f.Name())
}

func MustCallFrameFromProto(pb *pbprogram.CallFrame) *CallFrame {
	f, err := CallFrameFromProto(pb)
	if err != nil {
		panic(err)
	}
	return f
}

func CallFrameFromProto(pb *pbprogram.CallFrame) (*CallFrame, error) {
	if err := pb.Validate(); err != nil {
		return nil, err
	}

	return (&CallFrame{pb: pb}).Clone(), nil
}

func NewCallFrame(name string, loc *Location) (*CallFrame, error) {
	return CallFrameFromProto(&pbprogram.CallFrame{
		Name: name,
		Loc:  loc.PB(),
	})
}

func MustNewCallFrame(name string, loc *Location) *CallFrame {
	f, err := NewCallFrame(name, loc)
	if err != nil {
		panic(err)
	}
	return f
}

func SprintCallStack(cs []*CallFrame) string {
	if cs == nil {
		return ""
	}

	var ls []string

	for i, f := range cs {
		ls = append(ls, fmt.Sprintf("  #%d %v", i, f))
	}

	return strings.Join(ls, "\n")
}
