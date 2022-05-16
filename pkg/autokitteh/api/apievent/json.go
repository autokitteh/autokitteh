package apievent

import (
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"

	pbevent "github.com/autokitteh/autokitteh/gen/proto/stubs/go/event"
)

var (
	_ json.Marshaler   = &Event{}
	_ json.Unmarshaler = &Event{}
)

func (e *Event) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(e.pb)
}

func (e *Event) UnmarshalJSON(bs []byte) error {
	if e.pb == nil {
		e.pb = &pbevent.Event{}
	}

	if err := protojson.Unmarshal(bs, e.pb); err != nil {
		return err
	}

	return e.pb.Validate()
}

//--

var (
	_ json.Marshaler   = &EventState{}
	_ json.Unmarshaler = &EventState{}
)

func (s *EventState) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(s.pb)
}

func (s *EventState) UnmarshalJSON(bs []byte) error {
	if s.pb == nil {
		s.pb = &pbevent.EventState{}
	}

	if err := protojson.Unmarshal(bs, s.pb); err != nil {
		return err
	}

	return s.pb.Validate()
}

//--

var (
	_ json.Marshaler   = &EventStateRecord{}
	_ json.Unmarshaler = &EventStateRecord{}
)

func (r *EventStateRecord) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(r.pb)
}

func (r *EventStateRecord) UnmarshalJSON(bs []byte) error {
	if r.pb == nil {
		r.pb = &pbevent.EventStateRecord{}
	}

	if err := protojson.Unmarshal(bs, r.pb); err != nil {
		return err
	}

	return r.pb.Validate()
}

//--

var (
	_ json.Marshaler   = &ProjectEventState{}
	_ json.Unmarshaler = &ProjectEventState{}
)

func (s *ProjectEventState) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(s.pb)
}

func (s *ProjectEventState) UnmarshalJSON(bs []byte) error {
	if s.pb == nil {
		s.pb = &pbevent.ProjectEventState{}
	}

	if err := protojson.Unmarshal(bs, s.pb); err != nil {
		return err
	}

	return s.pb.Validate()
}

//--

var (
	_ json.Marshaler   = &ProjectEventStateRecord{}
	_ json.Unmarshaler = &ProjectEventStateRecord{}
)

func (r *ProjectEventStateRecord) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(r.pb)
}

func (r *ProjectEventStateRecord) UnmarshalJSON(bs []byte) error {
	if r.pb == nil {
		r.pb = &pbevent.ProjectEventStateRecord{}
	}

	if err := protojson.Unmarshal(bs, r.pb); err != nil {
		return err
	}

	return r.pb.Validate()
}
