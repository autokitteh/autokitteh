package apilang

import (
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"

	pblang "github.com/autokitteh/autokitteh/api/gen/stubs/go/lang"
)

var (
	_ json.Marshaler   = &RunStateLogRecord{}
	_ json.Unmarshaler = &RunStateLogRecord{}
)

func (p *RunStateLogRecord) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(p.pb)
}

func (p *RunStateLogRecord) UnmarshalJSON(bs []byte) error {
	if p.pb == nil {
		p.pb = &pblang.RunStateLogRecord{}
	}

	if err := protojson.Unmarshal(bs, p.pb); err != nil {
		return err
	}

	return p.pb.Validate()
}

//--

var (
	_ json.Marshaler   = &RunState{}
	_ json.Unmarshaler = &RunState{}
)

func (p *RunState) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(p.pb)
}

func (p *RunState) UnmarshalJSON(bs []byte) error {
	if p.pb == nil {
		p.pb = &pblang.RunState{}
	}

	if err := protojson.Unmarshal(bs, p.pb); err != nil {
		return err
	}

	return p.pb.Validate()
}

//--

var (
	_ json.Marshaler   = &RunSummary{}
	_ json.Unmarshaler = &RunSummary{}
)

func (p *RunSummary) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(p.pb)
}

func (p *RunSummary) UnmarshalJSON(bs []byte) error {
	if p.pb == nil {
		p.pb = &pblang.RunSummary{}
	}

	if err := protojson.Unmarshal(bs, p.pb); err != nil {
		return err
	}

	return p.pb.Validate()
}
