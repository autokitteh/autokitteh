package apieventsrc

import (
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"

	pbeventsrc "gitlab.com/softkitteh/autokitteh/gen/proto/stubs/go/eventsrc"
)

var (
	_ json.Marshaler   = &EventSource{}
	_ json.Unmarshaler = &EventSource{}
)

func (s *EventSource) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(s.pb)
}

func (s *EventSource) UnmarshalJSON(bs []byte) error {
	if s.pb == nil {
		s.pb = &pbeventsrc.EventSource{}
	}

	if err := protojson.Unmarshal(bs, s.pb); err != nil {
		return err
	}

	return s.pb.Validate()
}

//--

var (
	_ json.Marshaler   = &EventSourceSettings{}
	_ json.Unmarshaler = &EventSourceSettings{}
)

func (s *EventSourceSettings) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(s.pb)
}

func (s *EventSourceSettings) UnmarshalJSON(bs []byte) error {
	if s.pb == nil {
		s.pb = &pbeventsrc.EventSourceSettings{}
	}

	if err := protojson.Unmarshal(bs, s.pb); err != nil {
		return err
	}

	return s.pb.Validate()
}

//--

var (
	_ json.Marshaler   = &EventSourceProjectBinding{}
	_ json.Unmarshaler = &EventSourceProjectBinding{}
)

func (p *EventSourceProjectBinding) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(p.pb)
}

func (p *EventSourceProjectBinding) UnmarshalJSON(bs []byte) error {
	if p.pb == nil {
		p.pb = &pbeventsrc.EventSourceProjectBinding{}
	}

	if err := protojson.Unmarshal(bs, p.pb); err != nil {
		return err
	}

	return p.pb.Validate()
}

//--

var (
	_ json.Marshaler   = &EventSourceProjectBindingSettings{}
	_ json.Unmarshaler = &EventSourceProjectBindingSettings{}
)

func (p *EventSourceProjectBindingSettings) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(p.pb)
}

func (p *EventSourceProjectBindingSettings) UnmarshalJSON(bs []byte) error {
	if p.pb == nil {
		p.pb = &pbeventsrc.EventSourceProjectBindingSettings{}
	}

	if err := protojson.Unmarshal(bs, p.pb); err != nil {
		return err
	}

	return p.pb.Validate()
}
