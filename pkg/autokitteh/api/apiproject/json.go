package apiproject

import (
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"

	pbproject "github.com/autokitteh/autokitteh/gen/proto/stubs/go/project"
)

var (
	_ json.Marshaler   = &Project{}
	_ json.Unmarshaler = &Project{}
)

func (p *Project) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(p.pb)
}

func (p *Project) UnmarshalJSON(bs []byte) error {
	if p.pb == nil {
		p.pb = &pbproject.Project{}
	}

	if err := protojson.Unmarshal(bs, p.pb); err != nil {
		return err
	}

	return p.pb.Validate()
}

//--

var (
	_ json.Marshaler   = &ProjectSettings{}
	_ json.Unmarshaler = &ProjectSettings{}
)

func (p *ProjectSettings) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(p.pb)
}

func (p *ProjectSettings) UnmarshalJSON(bs []byte) error {
	if p.pb == nil {
		p.pb = &pbproject.ProjectSettings{}
	}

	if err := protojson.Unmarshal(bs, p.pb); err != nil {
		return err
	}

	return p.pb.Validate()
}

//--

var (
	_ json.Marshaler   = &ProjectPlugin{}
	_ json.Unmarshaler = &ProjectPlugin{}
)

func (p *ProjectPlugin) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(p.pb)
}

func (p *ProjectPlugin) UnmarshalJSON(bs []byte) error {
	if p.pb == nil {
		p.pb = &pbproject.ProjectPlugin{}
	}

	if err := protojson.Unmarshal(bs, p.pb); err != nil {
		return err
	}

	return p.pb.Validate()
}
