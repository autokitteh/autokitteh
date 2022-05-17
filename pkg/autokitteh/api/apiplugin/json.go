package apiplugin

import (
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"

	pb "github.com/autokitteh/autokitteh/api/gen/stubs/go/plugin"
)

var (
	_ json.Marshaler   = &Plugin{}
	_ json.Unmarshaler = &Plugin{}
)

func (m *Plugin) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(m.pb)
}

func (m *Plugin) UnmarshalJSON(bs []byte) error {
	if m.pb == nil {
		m.pb = &pb.Plugin{}
	}

	if err := protojson.Unmarshal(bs, m.pb); err != nil {
		return err
	}

	return m.pb.Validate()
}

//--

var (
	_ json.Marshaler   = &PluginSettings{}
	_ json.Unmarshaler = &PluginSettings{}
)

func (m *PluginSettings) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(m.pb)
}

func (m *PluginSettings) UnmarshalJSON(bs []byte) error {
	if m.pb == nil {
		m.pb = &pb.PluginSettings{}
	}

	if err := protojson.Unmarshal(bs, m.pb); err != nil {
		return err
	}

	return m.pb.Validate()
}

//--

var (
	_ json.Marshaler   = &PluginMemberDesc{}
	_ json.Unmarshaler = &PluginMemberDesc{}
)

func (m *PluginMemberDesc) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(m.pb)
}

func (m *PluginMemberDesc) UnmarshalJSON(bs []byte) error {
	if m.pb == nil {
		m.pb = &pb.PluginMemberDesc{}
	}

	if err := protojson.Unmarshal(bs, m.pb); err != nil {
		return err
	}

	return m.pb.Validate()
}

//--

var (
	_ json.Marshaler   = &PluginDesc{}
	_ json.Unmarshaler = &PluginDesc{}
)

func (m *PluginDesc) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(m.pb)
}

func (m *PluginDesc) UnmarshalJSON(bs []byte) error {
	if m.pb == nil {
		m.pb = &pb.PluginDesc{}
	}

	if err := protojson.Unmarshal(bs, m.pb); err != nil {
		return err
	}

	return m.pb.Validate()
}

//--

var (
	_ json.Marshaler   = &PluginExecSettings{}
	_ json.Unmarshaler = &PluginExecSettings{}
)

func (m *PluginExecSettings) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(m.pb)
}

func (m *PluginExecSettings) UnmarshalJSON(bs []byte) error {
	if m.pb == nil {
		m.pb = &pb.PluginExecutionSettings{}
	}

	if err := protojson.Unmarshal(bs, m.pb); err != nil {
		return err
	}

	return m.pb.Validate()
}
