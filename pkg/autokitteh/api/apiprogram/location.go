package apiprogram

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	pbprogram "github.com/autokitteh/autokitteh/api/gen/stubs/go/program"
)

type Location struct{ pb *pbprogram.Location }

func (l *Location) PB() *pbprogram.Location { return proto.Clone(l.pb).(*pbprogram.Location) }
func (l *Location) Clone() *Location        { return &Location{pb: l.PB()} }

func (l *Location) String() (s string) {
	if l == nil || l.pb == nil {
		return
	}

	if p := l.Path(); p != nil {
		s = p.String()
	}

	if l.pb.Line == 0 && l.pb.Column == 0 {
		return
	}

	if s != "" {
		s += ":"
	}

	return fmt.Sprintf("%s%d.%d", s, l.pb.Line, l.pb.Column)
}

func LocationFromProto(pb *pbprogram.Location) (*Location, error) {
	if pb == nil {
		return nil, nil
	}

	if err := pb.Validate(); err != nil {
		return nil, err
	}

	return (&Location{pb: pb}).Clone(), nil
}

func MustLocationFromProto(pb *pbprogram.Location) *Location {
	l, err := LocationFromProto(pb)
	if err != nil {
		panic(err)
	}
	return l
}

func NewLocation(path *Path, line, col int32) (*Location, error) {
	return LocationFromProto(&pbprogram.Location{
		Path:   path.PB(),
		Line:   line,
		Column: col,
	})
}

func MustNewLocation(path *Path, line, col int32) *Location {
	l, err := NewLocation(path, line, col)
	if err != nil {
		panic(err)
	}
	return l
}

func (l *Location) Path() *Path {
	p, err := PathFromProto(l.pb.Path)
	if err != nil {
		panic(err)
	}

	return p
}
