package apiprogram

import (
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"

	pbprogram "github.com/autokitteh/autokitteh/api/gen/stubs/go/program"
)

var (
	_ json.Marshaler   = &Module{}
	_ json.Unmarshaler = &Module{}
)

func (m *Module) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(m.pb)
}

func (m *Module) UnmarshalJSON(bs []byte) error {
	if m.pb == nil {
		m.pb = &pbprogram.Module{}
	}

	if err := protojson.Unmarshal(bs, m.pb); err != nil {
		return err
	}

	return m.pb.Validate()
}

//--

var (
	_ json.Marshaler   = &Path{}
	_ json.Unmarshaler = &Path{}
)

func (p *Path) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(p.pb)
}

func (p *Path) UnmarshalJSON(bs []byte) error {
	if p.pb == nil {
		p.pb = &pbprogram.Path{}
	}

	if err := protojson.Unmarshal(bs, p.pb); err != nil {
		return err
	}

	return p.pb.Validate()
}

//--

var (
	_ json.Marshaler   = &Location{}
	_ json.Unmarshaler = &Location{}
)

func (l *Location) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(l.pb)
}

func (l *Location) UnmarshalJSON(bs []byte) error {
	if l.pb == nil {
		l.pb = &pbprogram.Location{}
	}

	if err := protojson.Unmarshal(bs, l.pb); err != nil {
		return err
	}

	return l.pb.Validate()
}

//--

var (
	_ json.Marshaler   = &Error{}
	_ json.Unmarshaler = &Error{}
)

func (e *Error) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(e.pb)
}

func (e *Error) UnmarshalJSON(bs []byte) error {
	if e.pb == nil {
		e.pb = &pbprogram.Error{}
	}

	if err := protojson.Unmarshal(bs, e.pb); err != nil {
		return err
	}

	return e.pb.Validate()
}
